#include <Arduino.h>

#include <RFID.h>
#include <Ultrassonic.h>
#include <Photoresistor.h>
#include <Stepper.h>
#include <MQTT.h>
#include <ESPAsyncWebServer.h>
#include <Preferences.h>
#include <ArduinoJson.h>
#include <Timer.h>
#include <web/index.h>
#include <web/wifi.h>
#include <ESPmDNS.h>
#include <HTTPClient.h>

void updateLockState(bool lock, String user = "system");


#define RFID_SDA_PIN 10
#define RFID_RST_PIN 9
#define RFID_SCK_PIN 12
#define RFID_MISO_PIN 13
#define RFID_MOSI_PIN 11
RFID rfid(RFID_SDA_PIN, RFID_RST_PIN);

// --- Ultrasonic
#define TRIGGER_PIN 16
#define ECHO_PIN 17
#define DISTANCE_THRESHOLD_CM 100
Ultrassonic ultrassonic(TRIGGER_PIN, ECHO_PIN, DISTANCE_THRESHOLD_CM);

// --- Light Sensor (LDR)
#define LDR_PIN 4
Photoresistor ldr(LDR_PIN);

// --- Authorized UID
byte authorizedUID[4] = {0xDE, 0xAD, 0xBE, 0xEF}; // Example UID, replace with actual

// --- MQTT
const char *MQTT_SERVER = "mqtt.raiiaa.dev";
const uint16_t MQTT_PORT = 1883;
MQTT mqtt(MQTT_SERVER, MQTT_PORT);

// --- Servers Config
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
AsyncWebServer server(80);

// --- Servers
Preferences preferences;

// --- Timer
// Timer autoCloseTimer(5000, []()
//                      { stepper.close(); }, Timer::Mode::OneShot);

// --- State
String lastUser = "none";
int failCount = 0;
bool isLocked = true;

Timer autoCloseTimer(5000, []()
                     { updateLockState(true, "auto_close"); }, Timer::Mode::OneShot);


void sendTelemetry(String eventType, String details, String rfidUid = "")
{
  StaticJsonDocument<300> doc;
  doc["device_id"] = "smartlock_esp32";
  doc["event"] = eventType;
  doc["details"] = details;

  if (rfidUid != "")
  {
    doc["rfid_uid"] = rfidUid;
  }

  doc["status"] = "";
  doc["distance_cm"] = ultrassonic.distance();
  doc["light_level"] = ldr.lightLevel();
  doc["fails"] = rfid.failCount();
  if (rfid.failCount() > 0)
  {
    doc["user"] = lastUser;
  }

  doc["rssi"] = WiFi.RSSI();
  doc["uptime"] = millis() / 1000; // Uptime in seconds

  char buffer[300];
  serializeJson(doc, buffer);
  mqtt.publish("lock/telemetry", buffer);
}

// --- STEP 1: Define the function FIRST ---
void updateLockState(bool lock, String user)
{
  isLocked = lock;
  sendTelemetry("status_change", isLocked ? "LOCKED" : "UNLOCKED");
}


void callback(char *topic, byte *payload, unsigned int length)
{
  String msg = "";
  for (int i = 0; i < length; i++)
  {
    msg += (char)payload[i];
  }
  Serial.print("[MQTT] Received message on topic ");
  Serial.print(topic);
  Serial.print(": ");
  Serial.println(msg);

  if (msg == "UNLOCK")
  {
    Serial.println("[MQTT] AI/Backend authorized door UNLOCK");
    updateLockState(false, "mqtt_command");
  }
  else if (msg == "LOCK")
  {
    Serial.println("[MQTT] AI/Backend command: LOCK");
    updateLockState(true, "mqtt_command");
  }
}

// --- Timer
Timer telemetryTimer(5000, []
                     { sendTelemetry("heartbeat", "Periodic status update"); }, Timer::Mode::Periodic);

// --- Setup
void setup()
{
  Serial.begin(115200);

  // Servers
  mqtt.subscribe("lock/commands", callback);
  mqtt.subscribe("lock/ai/response", callback);
  mqtt.setup();

  // Components
  // stepper.setup();
  ultrassonic.setup();
  ldr.setup();
  rfid.setup(RFID_SCK_PIN, RFID_MISO_PIN, RFID_MOSI_PIN);

  // Timers
  telemetryTimer.start();

  // Wifi
  preferences.begin("wifi", true);
  String storedSSID = preferences.getString("ssid", SSID);
  String storedPassword = preferences.getString("password", PASSWORD);
  preferences.end();

  int attempts = 0;

  WiFi.begin(storedSSID.c_str(), storedPassword.c_str());
  while (WiFi.status() != WL_CONNECTED && attempts < 20)
  {
    delay(500);
    Serial.print(".");
    attempts++;
  }
  if (WiFi.status() != WL_CONNECTED)
  {
    Serial.println("\nWiFi connection failed! Continuing without WiFi...");
  }
  else
  {
    Serial.println("\nWifi connected, IP: " + WiFi.localIP().toString());
  }


  // mDNS
  if (!MDNS.begin("smartlock"))
  {
    Serial.println("Error setting up MDNS responder!");
  }

  // Web Server
  server.on("/", HTTP_GET, [](AsyncWebServerRequest *request)
            { request->send(200, "text/html", INDEX_HTML); });
  server.on("/wifi", HTTP_GET, [](AsyncWebServerRequest *request)
            { request->send(200, "text/html", WIFI_HTML); });
  server.on("/open", HTTP_GET, [](AsyncWebServerRequest *request) {
    updateLockState(false, "web_button");
    autoCloseTimer.reset();
    autoCloseTimer.start();
    request->send(200, "text/plain", "PORTA ABERTA");
  });
  server.on("/status", HTTP_GET, [](AsyncWebServerRequest *request) {
    request->send(200, "text/plain", isLocked ? "LOCKED" : "UNLOCKED");
  });
  server.on("/wifi-save", HTTP_POST, [](AsyncWebServerRequest *request) {
    if (request->hasParam("ssid", true) && request->hasParam("password", true)) {
      String ssid = request->getParam("ssid", true)->value();
      String password = request->getParam("password", true)->value();

      preferences.begin("wifi", false);
      preferences.putString("ssid", ssid);
      preferences.putString("password", password);
      preferences.end();

      request->send(200, "text/plain", "OK");

      delay(1000);
      ESP.restart();
    } else {
      request->send(400, "text/plain", "Missing ssid or password");
    }
  });
  server.on("/wifi-info", HTTP_GET, [](AsyncWebServerRequest *request) {
    request->send(200, "text/plain", "IP: " + WiFi.localIP().toString());
  });
  server.on("/users", HTTP_GET, [](AsyncWebServerRequest *request) {
    HTTPClient http;
    http.begin("https://smartlock-api.raiiaa.dev/api/users");
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      request->send(200, "application/json", payload);
    } else {
      request->send(500, "text/plain", "[]");
    }
    http.end();
  });
  server.on("/user-details", HTTP_GET, [](AsyncWebServerRequest *request) {
    if (!request->hasParam("uid")) {
      request->send(400, "text/plain", "Missing uid");
      return;
    }
    String uid = request->getParam("uid")->value();
    HTTPClient http;
    http.begin("https://smartlock-api.raiiaa.dev/api/users/" + uid);
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      request->send(200, "application/json", payload);
    } else {
      request->send(500, "text/plain", "{}");
    }
    http.end();
  });
  server.on("/check-services", HTTP_GET, [](AsyncWebServerRequest *request) {
    HTTPClient http;
    http.begin("https://smartlock-api.raiiaa.dev/api/health");
    int httpCode = http.GET();
    String backendStatus = "{}";
    if (httpCode > 0) {
      backendStatus = http.getString();
    }
    http.end();

    String response = "{\"local_mqtt\":" + String(mqtt.connected() ? "true" : "false") +
                      ",\"backend_services\":" + backendStatus + "}";
    request->send(200, "application/json", response);
  });
  server.begin();

}

void loop()
{
  // Servers
  mqtt.update();

  // Components
  // stepper.update();
  ultrassonic.update();
  ldr.update();
  rfid.update();

  // Timers
  telemetryTimer.update();
  autoCloseTimer.update();


  // --- 3. RFID Check (polled every 100ms for stable RF field and responsive scan detection)
  static unsigned long lastRFIDCheck = 0;
  if (millis() - lastRFIDCheck >= 100)
  {
    lastRFIDCheck = millis();
    if (rfid.check(authorizedUID))
    {
      lastUser = "authorized_user"; // In real case, map UID to user
      updateLockState(false, lastUser);
      // autoCloseTimer.start();
      char uidStr[20];
      sprintf(uidStr, "%02X:%02X:%02X:%02X", rfid.lastMatchUID()[0], rfid.lastMatchUID()[1], rfid.lastMatchUID()[2], rfid.lastMatchUID()[3]);
      sendTelemetry("access_granted", "Valid RFID", uidStr);
    }
    else
    {
      lastUser = "unknown";
      char uidStr[20];
      sprintf(uidStr, "%02X:%02X:%02X:%02X", rfid.buffer()[0], rfid.buffer()[1], rfid.buffer()[2], rfid.buffer()[3]);
      sendTelemetry("access_denied", "Invalid RFID. Fails: " + String(rfid.failCount()), uidStr);
    }
  }

  // --- Periodic Status Print
  static unsigned long lastStatusPrint = 0;
  unsigned long currentMillis = millis();
  if (currentMillis - lastStatusPrint >= 5000)
  {
    lastStatusPrint = currentMillis;
    Serial.print("[Status] Uptime: ");
    Serial.print(millis() / 1000);
    Serial.print("s | Distance: ");
    Serial.print(ultrassonic.distance());
    Serial.print(" cm | Light: ");
    Serial.print(ldr.lightLevel());
    Serial.print(" | RFID: ");
    if (rfid.isConnected()) {
      Serial.print("CONNECTED (v0x");
      Serial.print(rfid.getVersion(), HEX);
      Serial.print(")");
    } else {
      Serial.print("DISCONNECTED");
    }
    Serial.print(" | Lock: ");
    Serial.println(isLocked ? "LOCKED" : "UNLOCKED");
  }

  delay(10); // Reduce CPU usage slightly, allowing faster RFID polling
}
