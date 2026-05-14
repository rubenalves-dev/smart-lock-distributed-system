#include <MFRC522.h>

class RFID
{
private:
    int sdaPin_;
    int rstPin_;
    MFRC522 mfrc522_;
    uint failCount_ = 0;
    byte lastMatchUID_[4] = {0, 0, 0, 0};

public:
    RFID(int sdaPin, int rstPin);
    void setup();
    bool check(byte uid[4]);
    void update();
    void halt();

    uint failCount() const { return failCount_; }
    byte *lastMatchUID() { return lastMatchUID_; }
};
