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
#include "components/stepper.h"
#include "components/ultrassonic.h"
#include "components/photoresistor.h"
#include "components/rfid.h"

// --- RFID
#define RFID_SDA_PIN 21
#define RFID_RST_PIN 22
RFID rfid(RFID_SDA_PIN, RFID_RST_PIN);

// --- Ultrasonic
#define TRIGGER_PIN 5
#define ECHO_PIN 17
#define DISTANCE_THRESHOLD_CM 100
Ultrassonic ultrassonic(TRIGGER_PIN, ECHO_PIN, DISTANCE_THRESHOLD_CM);

// --- Light Sensor (LDR)
#define LDR_PIN 34
Photoresistor ldr(LDR_PIN);

// --- Stepper Motor
#define STEPPER_PIN_1 32
#define STEPPER_PIN_2 33
#define STEPPER_PIN_3 25
#define STEPPER_PIN_4 26
Stepper stepper(STEPPER_PIN_1, STEPPER_PIN_2, STEPPER_PIN_3, STEPPER_PIN_4);

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
    doc["details"] = details;

    doc["status"] = stepper.stateString();
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

// --- Timer
Timer telemetryTimer(5000, []
                     { sendTelemetry("heartbeat", "Periodic status update"); }, Timer::Mode::Periodic);

// --- Setup
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

    // Components
    stepper.setup();
    ultrassonic.setup();
    ldr.setup();
    rfid.setup();

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
    // MQTT Loop
    if (!mqttClient.connected())
    {
        reconnect();
    }
    mqttClient.loop();

    // Components
    stepper.update();
    ultrassonic.update();
    ldr.update();
    rfid.update();

    // Timers
    telemetryTimer.update();

    // --- 3. RFID Check
    if (ultrassonic.isObjectClose() && stepper.state() == Stepper::State::Closed)
    {
        if (rfid.check(authorizedUID))
        {
            lastUser = "authorized_user"; // In real case, map UID to user
            updateLockState(false, lastUser);
            // autoCloseTimer.start();
            sendTelemetry("access_granted", "Valid RFID");
        }
        else
        {
            lastUser = "unknown";
            sendTelemetry("access_denied", "Invalid RFID");
        }
    }

    delay(1000);
}
