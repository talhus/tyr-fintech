package worker

import (
	"encoding/json"
	"fmt"

	"github.com/iamtbay/tyr-fintech/internal/notifications"
	"github.com/iamtbay/tyr-fintech/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventConsumer struct {
	conn *amqp.Connection
	hub  *notifications.Hub
}

func NewEventConsumer(conn *amqp.Connection, hub *notifications.Hub) *EventConsumer {
	return &EventConsumer{
		conn: conn,
		hub:  hub,
	}
}

func (c *EventConsumer) StartConsuming() error {
	//open consumer ch
	ch, err := c.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open consumer channel")
	}

	// declare queue
	q, err := ch.QueueDeclare(
		"notifications_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// bind the queue
	err = ch.QueueBind(
		q.Name,
		"#",
		queue.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	// start receveing msg channel delivery stream from rabbitMQ
	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming")
	}

	fmt.Println("[RabbitMQ Worker] Succesfully subscribed to queue 'notifications_queue',Waiting")

	//go routine to consume msgs
	go func() {
		for d := range msgs {
			var e queue.EventPayload
			err := json.Unmarshal(d.Body, &e)
			if err != nil {
				fmt.Printf("Error marshaling RabbitMQ message %v \n", err)
				d.Nack(false, false)
				continue
			}

			fmt.Printf("[RabbitMQ Worker] Consumed message Type: %s, UserID: %s \n", e.EventType, e.UserID)

			notificationMsg := fmt.Sprintf("event: %s - Data: %v", e.EventType, e.Data)
			c.hub.SendToUser(e.UserID, notificationMsg)
			d.Ack(false)
		}
	}()

	return nil
}
