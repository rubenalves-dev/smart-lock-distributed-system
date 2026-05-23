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
	mu      sync.RWMutex
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   amqp.Queue
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
					log.Println("Connected to RabbitMQ successfully")
					return &RabbitMQClient{conn: conn, channel: ch, queue: q}, nil
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
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
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
		"",           // exchange
		r.queue.Name, // routing key
		false,        // mandatory
		false,        // immediate
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

	r.conn = conn
	r.channel = ch
	r.queue = q
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
