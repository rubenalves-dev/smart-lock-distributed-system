#include <AccelStepper.h>
#include "Stepper.h"

Stepper::Stepper(int stepPin1, int stepPin2, int stepPin3, int stepPin4)
    : stepper_(AccelStepper::FULL4WIRE, stepPin1, stepPin2, stepPin3, stepPin4)
{
}

void Stepper::setup()
{
    stepper_.setMaxSpeed(MAX_SPEED);
    stepper_.setAcceleration(ACCELERATION);
}

void Stepper::update()
{
    if (state_ == State::Opening || state_ == State::Closing)
    {
        if (!stepper_.run())
        { // When run() returns false, movement complete
            state_ = (state_ == State::Opening) ? State::Open : State::Closed;
            if (state_ == State::Open)
            {
                Serial.println("[LOCK] Open");
            }
            else
            {
                Serial.println("[LOCK] Closed");
            }
        }
    }
}

void Stepper::open(int dPin)
{
    stepper_.move(512);
    state_ = State::Opening;
    Serial.println("[LOCK] Opening...");
    while (stepper_.run())
    {
        if (dPin != -1)
        {
            digitalWrite(dPin, HIGH);
        }
    }
    state_ = State::Open;
    Serial.println("[LOCK] Open");
}

void Stepper::close(int dPin)
{
    stepper_.move(-512);
    state_ = State::Closing;
    Serial.println("[LOCK] Closing...");
    while (stepper_.run())
    {
        if (dPin != -1)
        {
            digitalWrite(dPin, HIGH);
        }
    }
    state_ = State::Closed;
    Serial.println("[LOCK] Closed");
}

const char *Stepper::stateString() const
{
    switch (state_)
    {
    case State::Closed:
        return "CLOSED";
    case State::Opening:
        return "OPENING";
    case State::Closing:
        return "CLOSING";
    case State::Open:
        return "OPEN";
    default:
        return "UNKNOWN";
    }
}
