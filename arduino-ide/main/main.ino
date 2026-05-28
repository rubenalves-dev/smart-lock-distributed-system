#include <Arduino.h>
#include <ArduinoJson.h>
#include <ESPmDNS.h>
#include <Preferences.h>
#include <WebServer.h>
#include <HTTPClient.h>
#include <WiFi.h>

#include <INDEX.h>
#include "WIFI_HTML.h"

#include "Timer.h"
#include "RFID.h"
#include "Ultrassonic.h"
#include "Photoresistor.h"
#include "MQTT.h"

void updateLockState(bool lock, String user = "system");

// --- RFID
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
Photoresistor ldr(LDR_PIN, 5000);

// --- MQTT
const char *MQTT_SERVER = "api.smartlock.raiiaa.dev";
const uint16_t MQTT_PORT = 1883;
MQTT mqtt(MQTT_SERVER, MQTT_PORT);

// --- Authorized UID
byte authorizedUID[4] = { 0xDE, 0xAD, 0xBE, 0xEF };  // Example UID, replace with actual

// --- Servers Config
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
WebServer server(80);

// --- State
String lastUser = "none";
int failCount = 0;
bool isLocked = true;

// --- Servers
Preferences preferences;

void sendTelemetry(String eventType, String details, String rfidUid = "");

// --- Timers
Timer autoCloseTimer(5000, []() { updateLockState(true, "auto_close"); }, Timer::Mode::OneShot);
Timer telemetryTimer(5000, []() { sendTelemetry("heartbeat", "Periodic status update"); }, Timer::Mode::Periodic);

// --- Functions
void sendTelemetry(String eventType, String details, String rfidUid)
{
  if (WiFi.status() != WL_CONNECTED)
  {
    return;
  }
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
  Serial.print("Received MQTT message: ");
  Serial.println(msg);

  if (msg == "UNLOCK")
  {
    updateLockState(false);
  }
  else if (msg == "LOCK")
  {
    updateLockState(true);
  }
}

void setupWebServer()
{
  server.on("/", HTTP_GET, []() {
    server.send(200, "text/html", INDEX_HTML);
  });
  server.on("/wifi", HTTP_GET, []() {
    server.send(200, "text/html", WIFI_HTML);
  });
  server.on("/open", HTTP_GET, []() {
    updateLockState(false, "web_button");
    autoCloseTimer.reset();
    autoCloseTimer.start();
    server.send(200, "text/plain", "PORTA ABERTA");
  });
  server.on("/status", HTTP_GET, []() {
    server.send(200, "text/plain", isLocked ? "LOCKED" : "UNLOCKED");
  });
  server.on("/wifi-save", HTTP_POST, []() {
    if (server.hasArg("ssid") && server.hasArg("password")) {
      String ssid = server.arg("ssid");
      String password = server.arg("password");

      preferences.begin("wifi", false);
      preferences.putString("ssid", ssid);
      preferences.putString("password", password);
      preferences.end();

      server.send(200, "text/plain", "OK");

      delay(1000);
      ESP.restart();
    } else {
      server.send(400, "text/plain", "Missing ssid or password");
    }
  });
  server.on("/wifi-info", HTTP_GET, []() {
    if (WiFi.status() == WL_CONNECTED) {
      server.send(200, "text/plain", "IP: " + WiFi.localIP().toString());
    } else {
      server.send(200, "text/plain", "IP: " + WiFi.softAPIP().toString() + " (AP Mode)");
    }
  });
  server.on("/users", HTTP_GET, []() {
    if (WiFi.status() != WL_CONNECTED) {
      server.send(500, "text/plain", "[]");
      return;
    }
    HTTPClient http;
    http.begin("https://api.smartlock.raiiaa.dev/api/users");
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      server.send(200, "application/json", payload);
    } else {
      server.send(500, "text/plain", "[]");
    }
    http.end();
  });
  server.on("/user-details", HTTP_GET, []() {
    if (WiFi.status() != WL_CONNECTED) {
      server.send(500, "text/plain", "{}");
      return;
    }
    if (!server.hasArg("uid")) {
      server.send(400, "text/plain", "Missing uid");
      return;
    }
    String uid = server.arg("uid");
    HTTPClient http;
    http.begin("https://api.smartlock.raiiaa.dev/api/users/" + uid);
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      server.send(200, "application/json", payload);
    } else {
      server.send(500, "text/plain", "{}");
    }
    http.end();
  });
  server.on("/check-services", HTTP_GET, []() {
    String backendStatus = "{}";
    if (WiFi.status() == WL_CONNECTED) {
      HTTPClient http;
      http.begin("https://api.smartlock.raiiaa.dev/api/health");
      int httpCode = http.GET();
      if (httpCode > 0) {
        backendStatus = http.getString();
      }
      http.end();
    }

    String response = "{\"local_mqtt\":" + String(mqtt.connected() ? "true" : "false") +
                      ",\"backend_services\":" + backendStatus + "}";
    server.send(200, "application/json", response);
  });
  server.begin();
}

void setup() {
  Serial.begin(115200);

  // Components
  ultrassonic.setup();
  ldr.setup();
  rfid.setup(RFID_SCK_PIN, RFID_MISO_PIN, RFID_MOSI_PIN);

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
    Serial.println("\nWiFi connection failed! Starting Access Point (AP mode)...");
    WiFi.softAP("SmartLock-Setup", "12345678");
    Serial.println("AP Mode started. SSID: SmartLock-Setup, Password: 12345678");
    Serial.print("Access the setup page at: http://");
    Serial.println(WiFi.softAPIP());
  }
  else
  {
    Serial.println("\nWifi connected, IP: " + WiFi.localIP().toString());
  }

  // Servers
  mqtt.subscribe("lock/commands", callback);
  mqtt.subscribe("lock/ai/response", callback);
  mqtt.setup();

  // Timers
  telemetryTimer.start();

  // mDNS
  if (!MDNS.begin("smartlock")) {
    Serial.println("Error setting up MDNS responder!");
  }

  // Web Server
  setupWebServer();
}

unsigned long lastSensorCheck = 0;

void loop() {
  server.handleClient();
  if (WiFi.status() == WL_CONNECTED) {
    mqtt.update();
  }
  telemetryTimer.update();
  autoCloseTimer.update();

  unsigned long currentMillis = millis();
  if (currentMillis - lastSensorCheck >= 1000) {
    lastSensorCheck = currentMillis;

    // Components
    ultrassonic.update();
    ldr.update();
    rfid.update();

    Serial.print(".");

    // --- 3. RFID Check
    if (ultrassonic.isObjectClose())
    {
      if (rfid.check(authorizedUID))
      {
        lastUser = "authorized_user"; // In real case, map UID to user
        updateLockState(false, lastUser);
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
  }

  delay(1);
}
