package core

import (
	"context"
	"fmt"
	"log"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)


type InfluxClient struct {
	Client influxdb2.Client
}

func NewInfluxClient(url, token string) (*InfluxClient, error) {
	client := influxdb2.NewClient(url, token)

	var err error
	for i := range 15 {
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

func (i *InfluxClient) QueryHealth(ctx context.Context, org, bucket, timeRange, interval string) ([]models.ServiceHealthSeries, error) {
	if i.Client == nil {
		return nil, fmt.Errorf("influx client is nil")
	}

	queryAPI := i.Client.QueryAPI(org)

	query := fmt.Sprintf(`from(bucket: "%s")
  |> range(start: -%s)
  |> filter(fn: (r) => r._measurement == "service_health")
  |> filter(fn: (r) => r._field == "status")
  |> aggregateWindow(every: %s, fn: max, createEmpty: true)`, bucket, timeRange, interval)

	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("influx query failed: %w", err)
	}
	defer result.Close()

	seriesMap := make(map[string][]models.HealthPoint)

	for result.Next() {
		record := result.Record()
		service := ""
		if sVal, ok := record.ValueByKey("service").(string); ok {
			service = sVal
		}
		if service == "" {
			continue
		}

		var statusVal *int
		if record.Value() != nil {
			if val, ok := record.Value().(int64); ok {
				status := int(val)
				statusVal = &status
			} else if val, ok := record.Value().(float64); ok {
				status := int(val)
				statusVal = &status
			}
		}

		pt := models.HealthPoint{
			Timestamp: record.Time().UTC().Format(time.RFC3339),
			Status:    statusVal,
		}

		seriesMap[service] = append(seriesMap[service], pt)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("error reading query results: %w", err)
	}

	var seriesList []models.ServiceHealthSeries
	for serviceName, points := range seriesMap {
		seriesList = append(seriesList, models.ServiceHealthSeries{
			Service: serviceName,
			Points:  points,
		})
	}

	return seriesList, nil
}

