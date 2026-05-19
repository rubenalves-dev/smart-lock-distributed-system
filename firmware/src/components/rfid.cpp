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
    Serial.begin(115200);
    SPI.begin();
    mfrc522_.PCD_Init();
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

    bool match = memcmp(mfrc522_.uid.uidByte, uid, mfrc522_.uid.size) == 0;
    halt();

    buffer_[0] = mfrc522_.uid.uidByte[0];
    buffer_[1] = mfrc522_.uid.uidByte[1];
    buffer_[2] = mfrc522_.uid.uidByte[2];
    buffer_[3] = mfrc522_.uid.uidByte[3];

    if (!match)
    {
        failCount_ = 0;
        lastMatchUID_[0] = mfrc522_.uid.uidByte[0];
        lastMatchUID_[1] = mfrc522_.uid.uidByte[1];
        lastMatchUID_[2] = mfrc522_.uid.uidByte[2];
        lastMatchUID_[3] = mfrc522_.uid.uidByte[3];
    }
    else
    {
        failCount_++;
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
