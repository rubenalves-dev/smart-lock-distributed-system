package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	mu             sync.RWMutex
	conn           *amqp.Connection
	channel        *amqp.Channel
	sensorQueue    amqp.Queue
	heartbeatQueue amqp.Queue
}

// NewRabbitMQClient initializes the connection with retries
func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	var conn *amqp.Connection
	var err error

	for i := range 15 {
		conn, err = amqp.Dial(url)
		if err == nil {
			ch, err := conn.Channel()
			if err == nil {
				q, err := ch.QueueDeclare(
					"sensor_events", // name
					true,            // durable
					false,           // delete when unused
					false,           // exclusive
					false,           // no-wait
					nil,             // arguments
				)
				if err == nil {
					hq, err := ch.QueueDeclare(
						"heartbeat_events", // name
						true,               // durable
						false,
						false,
						false,
						nil,
					)
					if err == nil {
						log.Println("Connected to RabbitMQ successfully")
						return &RabbitMQClient{conn: conn, channel: ch, sensorQueue: q, heartbeatQueue: hq}, nil
					}
				}
				ch.Close()
			}
			conn.Close()
		}
		log.Printf("Failed to connect to RabbitMQ (URL: %s), retrying in 2 seconds... (%d/15): %v", url, i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to RabbitMQ after retries: %w", err)
}

// PublishSensorEvent sends data to the AI service asynchronously
func (r *RabbitMQClient) PublishSensorEvent(body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.channel == nil || r.conn == nil || r.conn.IsClosed() {
		return amqp.ErrClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(ctx,
		"",                  // exchange
		r.sensorQueue.Name,  // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

// PublishHeartbeat sends heartbeat data asynchronously
func (r *RabbitMQClient) PublishHeartbeat(body []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.channel == nil || r.conn == nil || r.conn.IsClosed() {
		return amqp.ErrClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(ctx,
		"",                     // exchange
		r.heartbeatQueue.Name,  // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
}

func (r *RabbitMQClient) PublishRequestMFA() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.channel == nil || r.conn == nil || r.conn.IsClosed() {
		return amqp.ErrClosed
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.channel.PublishWithContext(ctx,
		"",                  // exchange
		r.sensorQueue.Name,  // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("request_mfa"),
		})
}

// IsConnected checks if the connection is active and not closed
func (r *RabbitMQClient) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.conn != nil && !r.conn.IsClosed()
}

// Reconnect attempts to close any existing resources and establish a new connection
func (r *RabbitMQClient) Reconnect(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}

	q, err := ch.QueueDeclare(
		"sensor_events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}

	hq, err := ch.QueueDeclare(
		"heartbeat_events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return err
	}

	r.conn = conn
	r.channel = ch
	r.sensorQueue = q
	r.heartbeatQueue = hq
	return nil
}

// ConsumeHeartbeats consumes heartbeat messages and runs the handler asynchronously
func (r *RabbitMQClient) ConsumeHeartbeats(ctx context.Context, handler func([]byte) error) error {
	r.mu.RLock()
	ch := r.channel
	r.mu.RUnlock()

	if ch == nil {
		return amqp.ErrClosed
	}

	msgs, err := ch.Consume(
		"heartbeat_events",              // queue
		"go_backend_heartbeat_consumer", // consumer tag
		true,                            // auto-ack
		false,                           // exclusive
		false,                           // no-local
		false,                           // no-wait
		nil,                             // args
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Println("RabbitMQ heartbeat channel closed")
					return
				}
				if err := handler(msg.Body); err != nil {
					log.Printf("Error handling background heartbeat message: %v\n", err)
				}
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection and channel
func (r *RabbitMQClient) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.channel != nil {
		_ = r.channel.Close()
	}
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
