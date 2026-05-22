package core

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type InfluxClient struct {
	Client influxdb2.Client
}

func NewInfluxClient(url, token string) (*InfluxClient, error) {
	client := influxdb2.NewClient(url, token)

	var err error
	for i := 0; i < 15; i++ {
		ok, pingErr := client.Ping(context.Background())
		if pingErr == nil && ok {
			log.Println("Connected to InfluxDB successfully")
			return &InfluxClient{Client: client}, nil
		}
		err = pingErr
		if err == nil {
			err = fmt.Errorf("ping returned false")
		}
		log.Printf("Failed to connect to InfluxDB (URL: %s), retrying in 2 seconds... (%d/15): %v", url, i+1, err)
		time.Sleep(2 * time.Second)
	}

	client.Close()
	return nil, fmt.Errorf("failed to connect to InfluxDB after retries: %w", err)
}

func (i *InfluxClient) Ping(ctx context.Context) (bool, error) {
	if i.Client == nil {
		return false, fmt.Errorf("influx client is nil")
	}
	return i.Client.Ping(ctx)
}

func (i *InfluxClient) WriteAPI(org, bucket string) api.WriteAPI {
	return i.Client.WriteAPI(org, bucket)
}

func (i *InfluxClient) Close() {
	if i.Client != nil {
		i.Client.Close()
	}
}
