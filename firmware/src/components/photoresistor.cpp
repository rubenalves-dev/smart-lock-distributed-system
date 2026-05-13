#include <Arduino.h>
#include "photoresistor.h"

Photoresistor::Photoresistor(int pin, int threshold)
{
    pin_ = pin;
    threshold_ = threshold;
}

void Photoresistor::setup()
{
    pinMode(pin_, INPUT);
}

void Photoresistor::update()
{
    lightLevel_ = readLightLevel();
}

int Photoresistor::readLightLevel()
{
    return analogRead(pin_);
}
