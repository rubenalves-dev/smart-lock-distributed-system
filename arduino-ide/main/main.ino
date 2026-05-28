#include <Arduino.h>
#include <ArduinoJson.h>
#include <ESPmDNS.h>
#include <Preferences.h>

#include <INDEX.h>
#include <WIFI.h>

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
Photoresistor ldr(LDR_PIN);

// --- MQTT
const char *MQTT_SERVER = "host.wokwi.internal";
const uint16_t MQTT_PORT = 1883;
MQTT mqtt(MQTT_SERVER, MQTT_PORT);

// --- Authorized UID
byte authorizedUID[4] = { 0xDE, 0xAD, 0xBE, 0xEF };  // Example UID, replace with actual

// --- Servers Config
// const char *SSID = "Wokwi-GUEST";
// const char *PASSWORD = "";
// AsyncWebServer server(80);

// --- Servers
Preferences preferences;

// --- Timers
Timer objectCloseTimer(
  500, [] {
    Serial.println("Object still close after 500ms");
  },
  Timer::Mode::Periodic);
Timer isLightTimer(
  500, [] {
    Serial.println("It's light after 500ms");
  },
  Timer::Mode::Periodic);

void setup() {
  Serial.begin(115200);

  ultrassonic.setup();
  ldr.setup();
  rfid.setup(RFID_SCK_PIN, RFID_MISO_PIN, RFID_MOSI_PIN);

  if (!MDNS.begin("smartlock")) {
    Serial.println("Error setting up MDNS responder!");
  }
}

void loop() {
  ultrassonic.update();
  ldr.update();
  rfid.update();

  Serial.print(".");

  if (ldr.isDark() && ultrassonic.isObjectClose()) {
    Serial.println();
    if (rfid.check(authorizedUID)) {
      Serial.println("Authorized user detected!");
    }
  }
}
