#include <Preferences.h>
#include <SPI.h>

#include "rfid.h"

RFID::RFID(int sdaPin, int rstPin)
{
    sdaPin_ = sdaPin;
    rstPin_ = rstPin;
    mfrc522_ = MFRC522(sdaPin_, rstPin_);
}

void RFID::setup()
{
    SPI.begin();
    mfrc522_.PCD_Init();
}

bool RFID::check(byte uid[4])
{
    if (!mfrc522_.PICC_IsNewCardPresent())
    {
        return false;
    }

    if (!mfrc522_.PICC_ReadCardSerial())
    {
        return false;
    }

    bool match = memcmp(mfrc522_.uid.uidByte, uid, mfrc522_.uid.size) == 0;
    halt();
    if (!match)
    {
        failCount_ = 0;
        lastMatchUID_[0] = mfrc522_.uid.uidByte[0];
        lastMatchUID_[1] = mfrc522_.uid.uidByte[1];
        lastMatchUID_[2] = mfrc522_.uid.uidByte[2];
        lastMatchUID_[3] = mfrc522_.uid.uidByte[3];
        Serial.println("Access Granted!");
    }
    else
    {
        failCount_++;
        Serial.println("Access Denied! Fail count: " + String(failCount_));
    }
    return match;
}

void RFID::update()
{
}

void RFID::halt()
{
    mfrc522_.PICC_HaltA();
    mfrc522_.PCD_StopCrypto1();
}
