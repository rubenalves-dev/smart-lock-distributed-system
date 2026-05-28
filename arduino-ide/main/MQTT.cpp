#include "MQTT.h"

MQTT::MQTT(const char *server, uint16_t port)
    : server_(server),
      port_(port),
      transport_(),
      client_(transport_)
{
}

int MQTT::state()
{
    return 0;
}

void MQTT::setup()
{
    client_.setServer(server_, port_);
    connect();
}

void MQTT::update()
{
    int attempts = 0;
    while (!connected())
    {
        connect();
        delay(1000);
        attempts++;
        if (attempts > 5)
        {
            Serial.println("MQTT connection failed after 5 attempts. Will retry in next loop.");
            return;
        }
    }

    client_.loop();
}

bool MQTT::connect()
{
    if (client_.connect("smartlock_esp32"))
    {
        Serial.println("MQTT connected");
        for (const auto &sub : subs_)
        {
            subscribe(sub.first, sub.second);
        }
        return true;
    }
    else
    {
        Serial.print("MQTT connection failed, rc=");
        Serial.println(client_.state());
        return false;
    }
}

bool MQTT::connected()
{
    return client_.connected();
}

bool MQTT::publish(const char *topic, const char *payload, bool retry)
{
    bool result = client_.publish(topic, payload);
    if (result)
    {
        return true;
    }

    if (connected())
    {
        Serial.println("MQTT publish failed but still connected. Will retry.");
        return false;
    }

    if (!connect())
    {
        Serial.println("MQTT publish failed and reconnection failed. Will retry in next loop.");
        return false;
    }

    if (retry)
    {
        Serial.println("MQTT publish retrying...");
        return publish(topic, payload, false);
    }

    return false;
}

bool MQTT::subscribe(const char *topic, MQTT_CALLBACK_SIGNATURE)
{
    client_.setCallback(callback);
    subs_.insert({topic, callback});
    return client_.subscribe(topic);
}
