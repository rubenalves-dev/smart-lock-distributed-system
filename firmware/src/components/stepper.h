#include <AccelStepper.h>

class Stepper
{
public:
    enum class State
    {
        Closed,
        Opening,
        Closing,
        Open
    };

private:
    const int MAX_SPEED = 500;
    const int ACCELERATION = 100;

    int pin1_;
    int pin2_;
    int pin3_;
    int pin4_;

    AccelStepper stepper_;
    State state_ = State::Closed;

public:
    Stepper(int pin1, int pin2, int pin3, int pin4);

    void setup();
    void update();
    void open(int dPin = -1);
    void close(int dPin = -1);

    State state() const { return state_; }
    const char *stateString() const;
};
