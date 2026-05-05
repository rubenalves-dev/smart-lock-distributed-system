#include <Arduino.h>
#include <ArduinoJson.h>
#include <Preferences.h>
#include <WiFi.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <PubSubClient.h>
#include <ESPmDNS.h>

#include "index.h"
#include "utils/timer.h"

// --- Servers Config ---
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
const char *MQTT_SERVER = "host.wokwi.internal";
const int MQTT_PORT = 1883;

// --- Pins ---
const int LOCK_PIN = 26;
const int STATUS_LED = 27;
const int WIFI_LED = 25;
const int SENSOR_PIN = 33;
const int VIBRATION_PIN = 14;

// --- Servers ---
AsyncWebServer server(80);
WiFiClient espClient;
PubSubClient mqttClient(espClient);
Preferences preferences;

// --- Variables ---
bool isLocked = true;
int failCount = 0;
String lastUser = "none";

// --- Timers ---
// Timer lockTimer(1000, []() {

// });

void sendTelemetry(String eventType, String details)
{
    if (!mqttClient.connected())
    {
        Serial.println("MQTT not connected, skipping telemetry");
        return;
    }

    StaticJsonDocument<300> doc;
    doc["device_id"] = "smartlock_esp32";
    doc["event"] = eventType;
    doc["status"] = isLocked ? "LOCKED" : "UNLOCKED";
    doc["user"] = lastUser;
    doc["fails"] = failCount;
    doc["rssi"] = WiFi.RSSI();
    doc["uptime"] = millis() / 1000;

    char buffer[300];
    serializeJson(doc, buffer);
    mqttClient.publish("lock/telemetry", buffer);
}

// --- STEP 1: Define the function FIRST ---
void updateLockState(bool lock, String user = "system")
{
    isLocked = lock;
    lastUser = user;
    digitalWrite(LOCK_PIN, isLocked ? LOW : HIGH);
    digitalWrite(STATUS_LED, isLocked ? LOW : HIGH);

    preferences.putBool("state", isLocked);
    sendTelemetry("status_change", isLocked ? "LOCKED" : "UNLOCKED");
}

void reconnect()
{
    while (!mqttClient.connected())
    {
        Serial.print("Connecting to MQTT...");
        if (mqttClient.connect("smartlock_esp32"))
        {
            Serial.println("connected");
            mqttClient.subscribe("lock/control");
        }
        else
        {
            Serial.print("failed, rc=");
            Serial.println(mqttClient.state());
            delay(5000);
        }
    }
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

void setup()
{
    Serial.begin(115200);
    pinMode(LOCK_PIN, OUTPUT);
    pinMode(STATUS_LED, OUTPUT);
    pinMode(WIFI_LED, OUTPUT);
    pinMode(SENSOR_PIN, INPUT_PULLUP);
    pinMode(VIBRATION_PIN, INPUT_PULLUP);

    preferences.begin("biometric", false);
    isLocked = preferences.getBool("state", true);
    updateLockState(isLocked);

    WiFi.begin(SSID, PASSWORD);
    while (WiFi.status() != WL_CONNECTED)
    {
        delay(500);
        Serial.print(".");
    }
    Serial.println("\nWifi connected, IP: " + WiFi.localIP().toString());
    analogWrite(WIFI_LED, 128);

    mqttClient.setServer(MQTT_SERVER, MQTT_PORT);
    mqttClient.setCallback(callback);

    server.on("/", HTTP_GET, [](AsyncWebServerRequest *request)
              { request->send(200, "text/html", INDEX_HTML); });
    server.on("/toggle", HTTP_GET, [](AsyncWebServerRequest *request)
              { updateLockState(!isLocked);
                request->send(200, "text/html", INDEX_HTML); });
    server.begin();
}

void loop()
{
    if (!mqttClient.connected())
    {
        reconnect();
    }
    mqttClient.loop();

    // Physical Button Logic
    if (digitalRead(SENSOR_PIN) == LOW)
    {
        Serial.println("Fingerprint Match!");
        failCount = 0;
        updateLockState(false, "um gajo random");
        delay(5000);
        updateLockState(true, "um gajo random");
    }

    // Vibration Sensor Logic
    if (digitalRead(VIBRATION_PIN) == LOW)
    {
        Serial.println("Vibration Detected! - Potential Breach!");
        failCount++;
        sendTelemetry("ALERT", "Vibration Detected");
        delay(1000);
    }
}
