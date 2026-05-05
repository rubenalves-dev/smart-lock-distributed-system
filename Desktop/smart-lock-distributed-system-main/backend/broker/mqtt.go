package broker

import (
	"encoding/json"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

func StartSubscriber(brokerIp string, dataChan chan<- models.SensorPayload) {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:1883", brokerIp))
	opts.SetClientID("go_backend_subscriber")

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("Connected to MQTT broker")
		c.Subscribe("lock/telemetry", 0, func(c mqtt.Client, m mqtt.Message) {
			var payload models.SensorPayload
			if err := json.Unmarshal(m.Payload(), &payload); err == nil {
				dataChan <- payload
			}
		})
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
