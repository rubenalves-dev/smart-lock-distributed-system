#include "timer.h"
#include <Arduino.h>

Timer::Timer(unsigned long duration, std::function<void()> callback, Mode mode)
{
    duration_ = duration;
    callback_ = callback;
    mode_ = mode;
    state_ = State::Idle;
    startTime_ = 0;
    pausedAt_ = 0;
    accumulatedPauseTime_ = 0;
}

void Timer::start()
{
    startTime_ = millis();
    accumulatedPauseTime_ = 0;
    state_ = State::Running;
}

void Timer::stop()
{
    state_ = State::Stopped;
}

void Timer::pause()
{
    if (state_ == State::Running)
    {
        pausedAt_ = millis();
        state_ = State::Paused;
    }
}

void Timer::resume()
{
    if (state_ == State::Paused)
    {
        accumulatedPauseTime_ += millis() - pausedAt_;
        state_ = State::Running;
    }
}

void Timer::reset()
{
    startTime_ = 0;
    pausedAt_ = 0;
    accumulatedPauseTime_ = 0;
    state_ = State::Idle;
}

void Timer::update()
{
    if (state_ != State::Running)
        return;

    unsigned long now = millis();
    unsigned long effectiveElapsed = now - startTime_ - accumulatedPauseTime_;

    if (effectiveElapsed >= duration_)
    {
        callback_();

        if (mode_ == Mode::Periodic)
        {
            startTime_ = millis();
            accumulatedPauseTime_ = 0;
        }
        else
        {
            state_ = State::Stopped;
        }
    }
}

Timer::State Timer::state() const
{
    return state_;
}

unsigned long Timer::elapsed() const
{
    if (state_ == State::Idle)
    {
        return 0;
    }

    unsigned long now = (state_ == State::Paused) ? pausedAt_ : millis();
    return now - startTime_ - accumulatedPauseTime_;
}

unsigned long Timer::remaining() const
{
    unsigned long e = elapsed();
    return (e >= duration_) ? 0 : (duration_ - e);
}

/*
EXAMPLE USAGE:

Timer myTimer(5000, []() {
    Serial.println("Timer expired!");
}, Timer::Mode::OneShot);

void setup() {
    Serial.begin(115200);
    myTimer.start();
}

void loop() {
    myTimer.update();
}
*/