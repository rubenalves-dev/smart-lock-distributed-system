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
const char *MQTT_SERVER = "host.wokwi.internal";
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
void updateLockState(bool lock, String user = "system")
{
  // if (lock)
  // {
  //   stepper.close();
  // }
  // else
  // {
  //   stepper.open();
  // }
  // sendTelemetry("status_change", stepper.stateString());
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
  int attempts = 0;

  WiFi.begin(SSID, PASSWORD);
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
  server.on("/toggle", HTTP_GET, [](AsyncWebServerRequest *request)
            { updateLockState(/* stepper.state() == Stepper::State::Closed ? false : true */true);
                request->send(200, "text/html", INDEX_HTML); });
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

  // --- 3. RFID Check
  if (ultrassonic.isObjectClose() /*&&stepper.state() == Stepper::State::Closed*/)
  {
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

  delay(1000); // Main loop delay to reduce CPU usage
}
