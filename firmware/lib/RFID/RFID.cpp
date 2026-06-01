#include <Preferences.h>
#include <SPI.h>

#include "RFID.h"

RFID::RFID(int sdaPin, int rstPin)
    : sdaPin_(sdaPin),
      rstPin_(rstPin),
      mfrc522_(sdaPin, rstPin)
{
}

void RFID::setup(int sckPin, int misoPin, int mosiPin)
{
    SPI.begin(sckPin, misoPin, mosiPin, sdaPin_);
    delay(100);
    mfrc522_.PCD_Init();
    delay(50);
    Serial.println("[RFID] Reader initialization triggered");
    byte v = getVersion();
    Serial.print("[RFID] MFRC522 Software Version: 0x");
    Serial.print(v, HEX);
    if (v == 0x91) {
        Serial.println(" = v1.0");
    } else if (v == 0x92) {
        Serial.println(" = v2.0");
    } else if (v == 0x88) {
        Serial.println(" = clone / FM17522");
    } else if (v == 0x00 || v == 0xFF) {
        Serial.println(" = WARNING: Communication failure! Check wiring.");
    } else {
        Serial.println(" = unknown version");
    }
}

bool RFID::check(byte uid[4])
{
    if (!mfrc522_.PICC_IsNewCardPresent())
    {
        Serial.println("No card present");
        return false;
    }

    if (!mfrc522_.PICC_ReadCardSerial())
    {
        Serial.println("Failed to read card serial");
        return false;
    }

    String cardId = "";
    for (byte i = 0; i < mfrc522_.uid.size; i++)
    {
        if (mfrc522_.uid.uidByte[i] < 0x10)
        {
            cardId += "0";
        }
        cardId += String(mfrc522_.uid.uidByte[i], HEX);
    }
    cardId.toUpperCase();
    Serial.println("Card UID: " + cardId);

    bool match = memcmp(mfrc522_.uid.uidByte, uid, mfrc522_.uid.size) == 0;
    halt();

    buffer_[0] = mfrc522_.uid.uidByte[0];
    buffer_[1] = mfrc522_.uid.uidByte[1];
    buffer_[2] = mfrc522_.uid.uidByte[2];
    buffer_[3] = mfrc522_.uid.uidByte[3];

    if (!match)
    {
        failCount_++;
        Serial.println("UID does not match authorized UID");
        Serial.println("Fail count: " + String(failCount_));
        Serial.print("Last read UID: ");
        for (byte i = 0; i < mfrc522_.uid.size; i++)
        {
            Serial.print(String(mfrc522_.uid.uidByte[i], HEX) + " ");
        }
        Serial.println();
    }
    else
    {
        failCount_ = 0;
        lastMatchUID_[0] = mfrc522_.uid.uidByte[0];
        lastMatchUID_[1] = mfrc522_.uid.uidByte[1];
        lastMatchUID_[2] = mfrc522_.uid.uidByte[2];
        lastMatchUID_[3] = mfrc522_.uid.uidByte[3];
    }
    return match;
}

void RFID::update()
{
}

byte RFID::getVersion() {
    return mfrc522_.PCD_ReadRegister(MFRC522::VersionReg);
}

bool RFID::isConnected() {
    byte v = getVersion();
    return (v != 0x00 && v != 0xFF);
}

void RFID::halt()
{
    mfrc522_.PICC_HaltA();
    mfrc522_.PCD_StopCrypto1();
}
