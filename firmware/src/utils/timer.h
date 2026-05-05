#pragma once
#include <functional>

class Timer
{
public:
    enum class State
    {
        Idle,
        Running,
        Paused,
        Stopped
    };
    enum class Mode
    {
        OneShot,
        Periodic
    };

    Timer(unsigned long duration, std::function<void()> callback, Mode mode = Mode::OneShot);

    void start();
    void stop();
    void pause();
    void resume();
    void reset();

    void update();

    State state() const;

    unsigned long elapsed() const;
    unsigned long remaining() const;

private:
    unsigned long duration_;
    unsigned long startTime_;
    unsigned long pausedAt_;
    unsigned long accumulatedPauseTime_;

    std::function<void()> callback_;

    Mode mode_;
    State state_;
};