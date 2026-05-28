#include <MFRC522.h>

class RFID
{
private:
    int sdaPin_;
    int rstPin_;
    MFRC522 mfrc522_;
    uint failCount_ = 0;
    byte lastMatchUID_[4] = {0, 0, 0, 0};
    byte buffer_[18]; // Buffer to hold data read from the card

public:
    RFID(int sdaPin, int rstPin);
    void setup(int sckPin, int misoPin, int mosiPin);
    bool check(byte uid[4]);
    bool is_new_uid_present();
    void update();
    void halt();

    uint failCount() const { return failCount_; }
    byte *lastMatchUID() { return lastMatchUID_; }
    byte *buffer() { return buffer_; }
};
