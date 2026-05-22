package monitor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/core"
)

type ServiceStatus struct {
	Service   string  `json:"service"`
	Online    bool    `json:"online"`
	LatencyMs float64 `json:"latency_ms"`
	Error     string  `json:"error,omitempty"`
}

type Monitor struct {
	db             *core.PostgresClient
	mqttClient     *core.MQTTClient
	influxClient   *core.InfluxClient
	rabbitClient   *core.RabbitMQClient
	rabbitURL      string
	influxOrg      string
	influxBucket   string
	mu             sync.RWMutex
	latestStatuses map[string]ServiceStatus
}

func NewMonitor(
	db *core.PostgresClient,
	mqttClient *core.MQTTClient,
	influxClient *core.InfluxClient,
	rabbitClient *core.RabbitMQClient,
	rabbitURL string,
	influxOrg string,
	influxBucket string,
) *Monitor {
	return &Monitor{
		db:             db,
		mqttClient:     mqttClient,
		influxClient:   influxClient,
		rabbitClient:   rabbitClient,
		rabbitURL:      rabbitURL,
		influxOrg:      influxOrg,
		influxBucket:   influxBucket,
		latestStatuses: make(map[string]ServiceStatus),
	}
}

// GetLatestStatuses returns a copy of the latest health statuses
func (m *Monitor) GetLatestStatuses() map[string]ServiceStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make(map[string]ServiceStatus)
	for k, v := range m.latestStatuses {
		statuses[k] = v
	}
	return statuses
}

// Start runs the monitoring loop every 10 seconds until context is done
func (m *Monitor) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Println("Background service health monitor started.")
	m.checkAndLogHealth(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("Service health monitor stopped.")
			return
		case <-ticker.C:
			m.checkAndLogHealth(ctx)
		}
	}
}

func (m *Monitor) checkAndLogHealth(ctx context.Context) {
	statuses := make(map[string]ServiceStatus)

	// 1. Check Postgres
	statuses["postgres"] = m.checkPostgres(ctx)

	// 2. Check Mosquitto
	statuses["mosquitto"] = m.checkMosquitto()

	// 3. Check RabbitMQ
	statuses["rabbitmq"] = m.checkRabbitMQ()

	// 4. Check InfluxDB
	statuses["influxdb"] = m.checkInfluxDB(ctx)

	// Store latest status in memory for API querying
	m.mu.Lock()
	m.latestStatuses = statuses
	m.mu.Unlock()

	// Write statuses to InfluxDB (if InfluxDB itself is online)
	if statuses["influxdb"].Online {
		m.writeStatusesToInflux(statuses)
	} else {
		log.Printf("Skipping InfluxDB write because InfluxDB is offline: %s", statuses["influxdb"].Error)
	}
}

func (m *Monitor) checkPostgres(ctx context.Context) ServiceStatus {
	start := time.Now()
	var err error
	if m.db != nil {
		err = m.db.Ping(ctx)
	} else {
		err = fmt.Errorf("postgres client is nil")
	}
	latency := time.Since(start).Seconds() * 1000.0

	status := ServiceStatus{
		Service:   "postgres",
		Online:    err == nil,
		LatencyMs: latency,
	}
	if err != nil {
		status.Error = err.Error()
	}
	return status
}

func (m *Monitor) checkMosquitto() ServiceStatus {
	start := time.Now()
	online := m.mqttClient != nil && m.mqttClient.IsConnected()
	latency := time.Since(start).Seconds() * 1000.0

	status := ServiceStatus{
		Service:   "mosquitto",
		Online:    online,
		LatencyMs: latency,
	}
	if !online {
		status.Error = "MQTT client is disconnected"
	}
	return status
}

func (m *Monitor) checkRabbitMQ() ServiceStatus {
	start := time.Now()
	var err error
	var online bool

	if m.rabbitClient != nil {
		online = m.rabbitClient.IsConnected()
		if !online {
			log.Println("RabbitMQ disconnected, attempting reconnection...")
			err = m.rabbitClient.Reconnect(m.rabbitURL)
			if err == nil {
				online = true
				log.Println("Successfully reconnected to RabbitMQ.")
			} else {
				log.Printf("Failed to reconnect to RabbitMQ: %v", err)
			}
		}
	} else {
		err = fmt.Errorf("rabbit client is nil")
	}

	latency := time.Since(start).Seconds() * 1000.0
	status := ServiceStatus{
		Service:   "rabbitmq",
		Online:    online,
		LatencyMs: latency,
	}
	if err != nil {
		status.Error = err.Error()
	} else if !online {
		status.Error = "RabbitMQ connection is closed"
	}
	return status
}

func (m *Monitor) checkInfluxDB(ctx context.Context) ServiceStatus {
	start := time.Now()
	var ok bool
	var err error
	if m.influxClient != nil {
		ok, err = m.influxClient.Ping(ctx)
	} else {
		err = fmt.Errorf("influx client is nil")
	}
	latency := time.Since(start).Seconds() * 1000.0

	status := ServiceStatus{
		Service:   "influxdb",
		Online:    ok && err == nil,
		LatencyMs: latency,
	}
	if err != nil {
		status.Error = err.Error()
	} else if !ok {
		status.Error = "InfluxDB ping returned false"
	}
	return status
}

func (m *Monitor) writeStatusesToInflux(statuses map[string]ServiceStatus) {
	if m.influxClient == nil {
		return
	}
	writeAPI := m.influxClient.WriteAPI(m.influxOrg, m.influxBucket)

	go func() {
		for err := range writeAPI.Errors() {
			log.Printf("InfluxDB Write error: %v", err)
		}
	}()

	for name, status := range statuses {
		statusVal := 0
		if status.Online {
			statusVal = 1
		}

		p := influxdb2.NewPoint(
			"service_health",
			map[string]string{"service": name},
			map[string]interface{}{
				"status":     statusVal,
				"latency_ms": status.LatencyMs,
				"error":      status.Error,
			},
			time.Now(),
		)
		writeAPI.WritePoint(p)
	}

	writeAPI.Flush()
}
