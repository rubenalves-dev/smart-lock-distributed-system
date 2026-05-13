class Photoresistor
{
private:
    int pin_;
    int threshold_;
    int lightLevel_ = 0;

public:
    Photoresistor(int pin, int threshold = 2000);

    void setup();
    void update();
    int readLightLevel();

    int lightLevel() const { return lightLevel_; }
    bool isDark() const
    {
        return lightLevel_ < threshold_;
    }
};