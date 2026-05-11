#include <Arduino.h>
#include <ArduinoJson.h>
#include <Preferences.h>
#include <WiFi.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <PubSubClient.h>
#include <ESPmDNS.h>

#include <SPI.h>
#include <MFRC522.h>

#include "index.h"
#include "utils/timer.h"
#include "components/stepper.h"

// --- RFID
#define RFID_SDA_PIN 21
#define RFID_RST_PIN 22
MFRC522 rfid(RFID_SDA_PIN, RFID_RST_PIN);

// --- Ultrasonic
#define TRIGGER_PIN 5
#define ECHO_PIN 18

// --- Light Sensor (LDR)
#define LDR_PIN 34

// --- Stepper Motor
#define STEPPER_SPEED 500
#define STEPPER_ACCEL 100
Stepper stepper(32, 33, 25, 26);

// --- LEDs
#define LED_OPENING 13
#define LED_OPEN 12
#define LED_CLOSING 2
#define LED_CLOSED 4

// --- Authorized UID
byte authorizedUID[4] = {0xDE, 0xAD, 0xBE, 0xEF}; // Example UID, replace with actual

// --- Servers Config
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
const char *MQTT_SERVER = "host.wokwi.internal";
const int MQTT_PORT = 1883;

// --- Servers
AsyncWebServer server(80);
WiFiClient espClient;
PubSubClient mqttClient(espClient);
Preferences preferences;

// --- Timer
// Timer autoCloseTimer(5000, []()
//                      { stepper.close(); }, Timer::Mode::OneShot);

// --- State
String lastUser = "none";
int failCount = 0;

// ---
void setStateLEDs()
{
    digitalWrite(LED_OPENING, stepper.state() == Stepper::State::Opening ? HIGH : LOW);
    digitalWrite(LED_OPEN, stepper.state() == Stepper::State::Open ? HIGH : LOW);
    digitalWrite(LED_CLOSING, stepper.state() == Stepper::State::Closing ? HIGH : LOW);
    digitalWrite(LED_CLOSED, stepper.state() == Stepper::State::Closed ? HIGH : LOW);
}

long readDistanceCm()
{
    digitalWrite(TRIGGER_PIN, LOW);
    delayMicroseconds(2);
    digitalWrite(TRIGGER_PIN, HIGH);
    delayMicroseconds(10);
    digitalWrite(TRIGGER_PIN, LOW);

    long duration = pulseIn(ECHO_PIN, HIGH, 30000);
    return duration * 0.034 / 2; // Convert to cm
}

bool checkRFID()
{
    if (!rfid.PICC_IsNewCardPresent() || !rfid.PICC_ReadCardSerial())
        return false;

    if (rfid.uid.size != sizeof(authorizedUID))
        return false;

    bool match = memcmp(rfid.uid.uidByte, authorizedUID, rfid.uid.size) == 0;

    rfid.PICC_HaltA();
    rfid.PCD_StopCrypto1();
    return match;
}

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
    doc["status"] = stepper.stateString();
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
    if (lock)
    {
        stepper.close();
    }
    else
    {
        stepper.open();
    }
    sendTelemetry("status_change", stepper.stateString());
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

    // LED pins
    pinMode(LED_OPENING, OUTPUT);
    pinMode(LED_OPEN, OUTPUT);
    pinMode(LED_CLOSING, OUTPUT);
    pinMode(LED_CLOSED, OUTPUT);
    setStateLEDs();

    // Ultrasonic
    pinMode(TRIGGER_PIN, OUTPUT);
    pinMode(ECHO_PIN, INPUT);

    // RFID
    SPI.begin();
    rfid.PCD_Init();

    // Stepper
    stepper.setup();

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

    // MQTT
    mqttClient.setServer(MQTT_SERVER, MQTT_PORT);
    mqttClient.setCallback(callback);

    // mDNS
    if (!MDNS.begin("smartlock"))
    {
        Serial.println("Error setting up MDNS responder!");
    }

    // Web Server
    server.on("/", HTTP_GET, [](AsyncWebServerRequest *request)
              { request->send(200, "text/html", INDEX_HTML); });
    server.on("/toggle", HTTP_GET, [](AsyncWebServerRequest *request)
              { updateLockState(stepper.state() == Stepper::State::Closed ? false : true);
                request->send(200, "text/html", INDEX_HTML); });
    server.begin();
}

void loop()
{
    Serial.println("\n--- LOOP START ---");
    // MQTT Loop
    if (!mqttClient.connected())
    {
        reconnect();
    }
    mqttClient.loop();

    // Stepper Loop
    stepper.run();

    // --- 1. Proximity Check
    long dist = readDistanceCm();
    bool personNearby = dist > 0 && dist < 100; // 100 cm threshold - 1 meter

    // --- 2. Light Level
    int lightLevel = analogRead(LDR_PIN);
    bool isDark = lightLevel < 2000; // Adjust threshold based on testing

    // --- 3. RFID Check
    if (personNearby && stepper.state() == Stepper::State::Closed)
    {
        if (checkRFID())
        {
            Serial.println("Access Granted!");
            lastUser = "authorized_user"; // In real case, map UID to user
            failCount = 0;
            updateLockState(false, lastUser);
            // autoCloseTimer.start();
            sendTelemetry("access_granted", "Valid RFID");
        }
        else
        {
            Serial.println("Access Denied!");
            lastUser = "unknown";
            failCount++;
            sendTelemetry("access_denied", "Invalid RFID");
        }
    }

    // --- Audit
    Serial.print("\nDistance (cm): ");
    Serial.print(dist);

    Serial.print("\nPerson Nearby: ");
    Serial.println(personNearby ? "Yes" : "No");

    Serial.print("\nLight Level (analog): ");
    Serial.println(lightLevel);

    Serial.print("\nLights On: ");
    Serial.println(isDark ? "No" : "Yes");

    Serial.print("\nLock State: ");
    Serial.println(stepper.stateString());

    Serial.println("\n--- LOOP END ---");
    delay(100);
}
