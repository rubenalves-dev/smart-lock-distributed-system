#include <Arduino.h>


// #include <include/web/wifi.h>
// #include <include/web/wifi.h>

#include "Timer.h"
#include "RFID.h"

void updateLockState(bool lock, String user = "system");

#define RFID_SDA_PIN 10
#define RFID_RST_PIN 9
#define RFID_SCK_PIN 12
#define RFID_MISO_PIN 13
#define RFID_MOSI_PIN 11
RFID rfid(RFID_SDA_PIN, RFID_RST_PIN);

void setup() {
  Serial.begin(115200);
  rfid.setup(RFID_SCK_PIN, RFID_MISO_PIN, RFID_MOSI_PIN);
}

void loop() {
  rfid.update();
  rfid.halt();
  delay(5000);
}
