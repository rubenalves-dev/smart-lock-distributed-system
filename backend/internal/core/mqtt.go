package core

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type MQTTClient struct {
	Client mqtt.Client
}

func NewMQTTClient(brokerIp string, dataChan chan<- models.SensorPayload) (*MQTTClient, error) {
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:1883", brokerIp))
	opts.SetClientID("go_backend_subscriber")
	opts.SetAutoReconnect(true)

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("Connected to MQTT broker")
		c.Subscribe("lock/telemetry", 0, func(c mqtt.Client, m mqtt.Message) {
			var payload models.SensorPayload
			if err := json.Unmarshal(m.Payload(), &payload); err == nil {
				dataChan <- payload
			} else {
				log.Printf("Failed to unmarshal MQTT message: %v\n", err)
			}
		})
	}

	client := mqtt.NewClient(opts)

	var err error
	for i := range 15 {
		token := client.Connect()
		if token.Wait() && token.Error() == nil {
			log.Println("Connected to MQTT broker successfully")
			return &MQTTClient{Client: client}, nil
		}
		err = token.Error()
		log.Printf("Failed to connect to MQTT broker (%s), retrying in 2 seconds... (%d/15): %v", brokerIp, i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to MQTT broker after retries: %w", err)
}

func (m *MQTTClient) PublishOpenDoor() error {
	token := m.Client.Publish("lock/control", 0, false, "open_door")
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) IsConnected() bool {
	return m.Client != nil && m.Client.IsConnected()
}

func (m *MQTTClient) Close() {
	if m.Client != nil && m.Client.IsConnected() {
		m.Client.Disconnect(250)
	}
}
