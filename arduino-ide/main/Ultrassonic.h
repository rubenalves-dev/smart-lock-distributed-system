class Ultrassonic
{
private:
    int triggerPin_;
    int echoPin_;
    double tresholdCm_;
    double distance_ = 0;

public:
    Ultrassonic(int triggerPin, int echoPin, double tresholdCm = 100.0);

    void setup();
    void update();
    double readDistanceCm();

    double distance() const { return distance_; }
    bool isObjectClose() const
    {
        return distance_ >= 0 && distance_ < tresholdCm_;
    }
};