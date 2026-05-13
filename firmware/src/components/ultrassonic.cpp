#include <Arduino.h>

#include "ultrassonic.h"

Ultrassonic::Ultrassonic(int triggerPin, int echoPin, double tresholdCm)
{
    triggerPin_ = triggerPin;
    echoPin_ = echoPin;
    tresholdCm_ = tresholdCm;
}

void Ultrassonic::setup()
{
    pinMode(triggerPin_, OUTPUT);
    pinMode(echoPin_, INPUT);
}

void Ultrassonic::update()
{
    distance_ = readDistanceCm();
}

double Ultrassonic::readDistanceCm()
{
    digitalWrite(triggerPin_, LOW);
    delayMicroseconds(2);
    digitalWrite(triggerPin_, HIGH);
    delayMicroseconds(10);
    digitalWrite(triggerPin_, LOW);

    long duration = pulseIn(echoPin_, HIGH, 30000);
    double distance = duration * 0.034 / 2;
    return distance; // Convert to cm
}

double Ultrassonic::distance() const
{
    return distance_;
}