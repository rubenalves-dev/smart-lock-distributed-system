#include <Arduino.h>
#include "index.h"
#include <WiFi.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <PubSubClient.h>
#include <Preferences.h>
#include <ESPmDNS.h>

// --- MQTT Config ---
const char *ssid = "Wokwi-GUEST";
const char *password = "";
const char *mqtt_server = "localhost:1883";

// --- Pins ---
const int LOCK_PIN = 26;
const int STATUS_LED = 27;
const int WIFI_LED = 25;
const int SENSOR_PIN = 33;
const int VIBRATION_PIN = 14;

// --- Objects ---
AsyncWebServer server(80);
WiFiClient espClient;
PubSubClient mqttClient(espClient);
Preferences preferences;

bool isLocked = true;

// --- STEP 1: Define the function FIRST ---
void updateLockState(bool lock)
{
    isLocked = lock;
    digitalWrite(LOCK_PIN, isLocked ? LOW : HIGH);
    digitalWrite(STATUS_LED, isLocked ? LOW : HIGH);

    // Save to EPROM
    preferences.putBool("state", isLocked);

    if (mqttClient.connected())
    {
        mqttClient.publish("lock/status", isLocked ? "LOCKED" : "UNLOCKED");
    }
}

// --- STEP 2: Define callback (it can now "see" updateLockState) ---
void callback(char *topic, byte *payload, unsigned int length)
{
    String message = "";
    for (int i = 0; i < length; i++)
        message += (char)payload[i];
    if (message == "UNLOCK")
        updateLockState(false);
    if (message == "LOCK")
        updateLockState(true);
}

// --- STEP 3: Define reconnect ---
void reconnect()
{
    while (!mqttClient.connected())
    {
        Serial.print("Connecting to MQTT...");
        if (mqttClient.connect("ESP32_Lock_Hugo_Final"))
        {
            Serial.println("connected");
            mqttClient.subscribe("lock/control");
        }
        else
        {
            Serial.print("failed, rc=");
            Serial.print(mqttClient.state());
            delay(5000);
        }
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

    WiFi.begin("Wokwi-GUEST", "");
    while (WiFi.status() != WL_CONNECTED)
    {
        delay(500);
    }
    analogWrite(WIFI_LED, 128);

    if (!MDNS.begin("fechadura"))
    {
        Serial.println("Erro ao configurar mDNS!");
    }
    else
    {
        Serial.println("mDNS configurado: http://fechadura.local");
    }

    mqttClient.setServer(mqtt_server, 1883);
    mqttClient.setCallback(callback);

    server.on("/", HTTP_GET, [](AsyncWebServerRequest *request)
              {
        if(request->hasParam("toggle")) {
            updateLockState(!isLocked);
            request->redirect("/");
        } else {
            request->send(200, "text/html", "Biometric Lock Status: " + String(isLocked ? "LOCKED" : "UNLOCKED"));
        } });

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
