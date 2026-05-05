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

// --- MQTT Config (CORRIGIDO PARA MAIÚSCULAS) ---
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
const char* MQTT_SERVER = "broker.hivemq.com"; // o docker fdd para mim
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

// --- Security Variables ---
bool isLocked = true;
int failCount = 0;           
String lastUser = "none";    


void sendTelemetry(String eventType, String details)
{
    if (!mqttClient.connected()) return;

    StaticJsonDocument<300> doc;
    doc["device_id"] = "ESP32_HUGO_01";
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


void updateLockState(bool lock, String user = "system")
{
    isLocked = lock;
    lastUser = user;
    digitalWrite(LOCK_PIN, isLocked ? LOW : HIGH);
    digitalWrite(STATUS_LED, isLocked ? LOW : HIGH);

    preferences.putBool("state", isLocked);
    sendTelemetry("status_change", isLocked ? "closed" : "opened");
}

void reconnect()
{
    while (!mqttClient.connected())
    {
        Serial.print("Connecting to MQTT...");
        // Tenta ligar com um ID único para evitar quedas de conexão
        if (mqttClient.connect("smartlock_hugo_esp32"))
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
    while (WiFi.status() != WL_CONNECTED) {
        delay(500);
        Serial.print(".");
    }
    Serial.println("\nWiFi Connected!");
    analogWrite(WIFI_LED, 128);

    // Agora o MQTT_SERVER está definido corretamente no topo
    mqttClient.setServer(MQTT_SERVER, MQTT_PORT);
    mqttClient.setCallback([](char* topic, byte* payload, unsigned int length) {
        String msg = "";
        for (int i = 0; i < length; i++) msg += (char)payload[i];
        if (msg == "UNLOCK") updateLockState(false, "remote_admin");
        if (msg == "LOCK") updateLockState(true, "remote_admin");
    });
}

void loop()
{
    if (!mqttClient.connected())
    {
        reconnect();
    }
    mqttClient.loop();

    // Lógica do Sensor de Impressão Digital (Simulado pelo SENSOR_PIN)
    if (digitalRead(SENSOR_PIN) == LOW)
    {
        Serial.println("Fingerprint Match!");
        failCount = 0; 
        updateLockState(false,"hugo");
        delay(5000);
        updateLockState(true, "hugo");
    }

    // Lógica do Sensor de Vibração (Simulado pelo VIBRATION_PIN)
    if (digitalRead(VIBRATION_PIN) == LOW)
    {
        Serial.println("Vibration Detected - Potential Breach!");
        failCount++; // Aumenta falhas para a IA detetar anomalia
        sendTelemetry("ALERT", "Vibration detected");
        delay(1000); // Debounce simples
    }
}