#include <map>
#include <PubSubClient.h>
#include <WiFiClient.h>

class MQTT
{
private:
    const char *server_;
    uint16_t port_;
    WiFiClient transport_;
    PubSubClient client_;

    std::map<const char *, std::function<void(char *, uint8_t *, unsigned int)>> subs_;

public:
    MQTT(const char *server, uint16_t port);

    int state();
    void setup();
    void update();

    bool connect();
    bool connected();
    bool publish(const char *topic, const char *payload, bool retry = true);
    bool subscribe(const char *topic, MQTT_CALLBACK_SIGNATURE);
};
