#include <Arduino.h>
#include <ArduinoJson.h>
#include <Preferences.h>
#include <WiFi.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <PubSubClient.h>
#include <ESPmDNS.h>

#include "components/stepper.h"
#include "components/ultrassonic.h"
#include "components/photoresistor.h"
#include "components/rfid.h"
#include "ports/mqtt.h"
#include "templates/index.h"
#include "templates/wifi.h"
#include "utils/timer.h"

// --- RFID
#define SDA_PIN 5
#define RST_PIN 6
#define SCK_PIN 18
#define MISO_PIN 19
#define MOSI_PIN 23
RFID rfid(SDA_PIN, RST_PIN);

// --- Ultrasonic
#define TRIGGER_PIN 16
#define ECHO_PIN 17
#define DISTANCE_THRESHOLD_CM 100
Ultrassonic ultrassonic(TRIGGER_PIN, ECHO_PIN, DISTANCE_THRESHOLD_CM);

// --- Light Sensor (LDR)
#define LDR_PIN 4
Photoresistor ldr(LDR_PIN);

// --- Stepper Motor
#define STEPPER_PIN_1 32
#define STEPPER_PIN_2 33
#define STEPPER_PIN_3 25
#define STEPPER_PIN_4 26
Stepper stepper(STEPPER_PIN_1, STEPPER_PIN_2, STEPPER_PIN_3, STEPPER_PIN_4);

// --- Authorized UID
byte authorizedUID[4] = {0xDE, 0xAD, 0xBE, 0xEF}; // Example UID, replace with actual

// // --- MQTT
// const char *MQTT_SERVER = "host.wokwi.internal";
// const uint16_t MQTT_PORT = 1883;
// MQTT mqtt(MQTT_SERVER, MQTT_PORT);

// // --- Servers Config
// const char *SSID = "Wokwi-GUEST";
// const char *PASSWORD = "";
// AsyncWebServer server(80);

// --- Servers
Preferences preferences;

// --- Timer
// Timer autoCloseTimer(5000, []()
//                      { stepper.close(); }, Timer::Mode::OneShot);

// --- State
String lastUser = "none";
int failCount = 0;

// void sendTelemetry(String eventType, String details)
// {
//     StaticJsonDocument<300> doc;
//     doc["device_id"] = "smartlock_esp32";
//     doc["event"] = eventType;
//     doc["details"] = details;

//     doc["status"] = stepper.stateString();
//     doc["distance_cm"] = ultrassonic.distance();
//     doc["light_level"] = ldr.lightLevel();
//     doc["fails"] = rfid.failCount();
//     if (rfid.failCount() > 0)
//     {
//         doc["user"] = lastUser;
//     }

//     doc["rssi"] = WiFi.RSSI();
//     doc["uptime"] = millis() / 1000; // Uptime in seconds

//     char buffer[300];
//     serializeJson(doc, buffer);
//     mqtt.publish("lock/telemetry", buffer);
// }

// // --- STEP 1: Define the function FIRST ---
// void updateLockState(bool lock, String user = "system")
// {
//     if (lock)
//     {
//         stepper.close();
//     }
//     else
//     {
//         stepper.open();
//     }
//     sendTelemetry("status_change", stepper.stateString());
// }

// void callback(char *topic, byte *payload, unsigned int length)
// {
//     String msg = "";
//     for (int i = 0; i < length; i++)
//     {
//         msg += (char)payload[i];
//     }
//     Serial.print("Received MQTT message: ");
//     Serial.println(msg);

//     if (msg == "UNLOCK")
//     {
//         updateLockState(false);
//     }
//     else if (msg == "LOCK")
//     {
//         updateLockState(true);
//     }
// }

// // --- Timer
// Timer telemetryTimer(5000, []
//                      { sendTelemetry("heartbeat", "Periodic status update"); }, Timer::Mode::Periodic);

// --- Setup
void setup()
{
    Serial.begin(115200);

    Serial.println("BOOT OK");

    SPI.begin(SCK_PIN, MISO_PIN, MOSI_PIN, SDA_PIN);

    rfid.setup(SCK_PIN, MISO_PIN, MOSI_PIN);

    Serial.println("RFID OK");
}

void loop()
{
    delay(1000);
}