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

// --- MQTT Config ---
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
const char *MQTT_SERVER = "192.168.1.219"; // TROQUEM PELO VOSSO IP PARA IR BUSCAR O MOSQUITTO DOCKER

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

// --- Timers ---
Timer lockTimer(1000, []() {

});

void sendTelemetry(String type, float value)
{
    if (!mqttClient.connected())
    {
        Serial.println("MQTT not connected, skipping telemetry");
        return;
    }

    StaticJsonDocument<200> doc;
    doc["device_id"] = "";
    doc["type"] = type;
    doc["value"] = value;
    doc["is_locked"] = isLocked;

    char buffer[200];
    serializeJson(doc, buffer);
    mqttClient.publish("lock/telemetry", buffer);
}

// --- STEP 1: Define the function FIRST ---
void updateLockState(bool lock)
{
    isLocked = lock;
    digitalWrite(LOCK_PIN, isLocked ? LOW : HIGH);
    digitalWrite(STATUS_LED, isLocked ? LOW : HIGH);

    preferences.putBool("state", isLocked);
    sendTelemetry("status_change", isLocked ? 1 : 0);
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
        delay(500);
    analogWrite(WIFI_LED, 128);

    MDNS.begin("fechadura");
    mqttClient.setServer(MQTT_SERVER, 1883);
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
        updateLockState(false);
        delay(5000);
        updateLockState(true);
    }

    // Vibration Sensor Logic
    if (digitalRead(VIBRATION_PIN) == LOW)
    {
        Serial.println("Vibration Detected!");
        updateLockState(false);
        delay(5000);
        updateLockState(true);
    }
}
