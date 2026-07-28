package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// standart json payload
type EventPayload struct {
	EventType string    `json:"event_type"`
	UserID    string    `json:"user_id"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type RabbitMQClient struct {
	conn *amqp.Connection

	channel *amqp.Channel
}

const ExchangeName = "fintech.events"

func NewRabbitMQ(url string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	Channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ Channel: %w", err)
	}

	//declare exchange
	err = Channel.ExchangeDeclare(
		ExchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // delete when unused
		false,        // no-wait
		false,        // internal
		nil,          // arguments
	)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to declare a RabbitMQ exchange: %w", err)
	}

	fmt.Println("Succesfully connected to RabbitMQ and declared exchange", ExchangeName)

	return &RabbitMQClient{
		conn:    conn,
		channel: Channel,
	}, nil
}

// PUBLISH EVENT
func (c *RabbitMQClient) PublishEvent(ctx context.Context, routingKey string, event EventPayload) error {
	//set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	//publish
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	//publish msg
	err = c.channel.PublishWithContext(ctx, ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
		Timestamp:    time.Now(),
	})

	if err != nil {
		return fmt.Errorf("Failed to publish event to RabbitMQ: %w", err)
	}

	fmt.Printf("[RabbitMQ] Published event '%s' with key '%s'\n", event.EventType, routingKey)
	return nil
}

func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}

func (r *RabbitMQClient) Conn() *amqp.Connection {
	return r.conn
}
