#include <Arduino.h>

#include "RFID.h"

// --- RFID
#define SDA_PIN 10
#define RST_PIN 9
#define SCK_PIN 12
#define MISO_PIN 13
#define MOSI_PIN 11
RFID rfid(SDA_PIN, RST_PIN);

void setup()
{
  // put your setup code here, to run once:
  Serial.begin(115200);
  rfid.setup(SCK_PIN, MISO_PIN, MOSI_PIN);
}

void loop()
{
  // put your main code here, to run repeatedly:
  rfid.update();
  Serial.print("Fail count: ");
  Serial.println(rfid.failCount());
  delay(1000);
}