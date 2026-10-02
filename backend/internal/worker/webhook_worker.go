package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamtbay/tyr-fintech/internal/dto"
	"github.com/iamtbay/tyr-fintech/internal/queue"
	amqp "github.com/rabbitmq/amqp091-go"
)

var WebHookQueue = make(chan *dto.TransactionWebhookEvent, 100)
var MerchantWebHookQueue = make(chan *dto.MerchantWebhookEvent, 100)

var globalRabbitMQClient *queue.RabbitMQClient

// InitRabbitMQDispatcher sets the RabbitMQ client for durable persistent dispatching
func InitRabbitMQDispatcher(client *queue.RabbitMQClient) {
	globalRabbitMQClient = client
}

// EnqueueMerchantWebhook routes the webhook event to durable RabbitMQ if available,
// while preserving in-memory channel delivery for test suites and offline fallbacks.
func EnqueueMerchantWebhook(event *dto.MerchantWebhookEvent) {
	// 1. In-memory channel non-blocking dispatch (keeps unit tests and fallbacks operational)
	select {
	case MerchantWebHookQueue <- event:
	default:
		log.Println("[Webhook Worker] In-memory MerchantWebHookQueue channel is full")
	}

	// 2. Publish to durable RabbitMQ topic if client is initialized
	if globalRabbitMQClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := globalRabbitMQClient.PublishMerchantWebhook(ctx, "webhook.merchant.event", event, 0); err != nil {
			log.Printf("[Webhook Worker] Failed to publish event to RabbitMQ: %v\n", err)
		} else {
			log.Printf("[Webhook Worker] Persisted event %s (Order: %s) to RabbitMQ queue\n", event.Event, event.OrderID)
		}
	}
}

// StartWebhookWorker starts the local background channel consumers
func StartWebhookWorker() {
	log.Println("Webhook worker has started, tasks are awaiting")
	go func() {
		for tx := range WebHookQueue {
			sendWebhook(tx)
		}
	}()

	// Only process in-memory channel directly if RabbitMQ is not connected
	go func() {
		for event := range MerchantWebHookQueue {
			if globalRabbitMQClient == nil {
				// Local fallback dispatch when RabbitMQ is offline
				sendMerchantWebhook(event, 1)
			}
		}
	}()
}

// StartMerchantWebhookConsumer connects a durable RabbitMQ consumer with exponential backoff & DLQ
func StartMerchantWebhookConsumer(conn *amqp.Connection) error {
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open merchant webhook channel: %w", err)
	}

	// Declare queue with Dead-Letter Exchange (DLX) routing
	args := amqp.Table{
		"x-dead-letter-exchange":    "fintech.events.dlx",
		"x-dead-letter-routing-key": "merchant_webhooks_dlq",
	}

	q, err := ch.QueueDeclare(
		"merchant_webhooks_queue",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		args,  // table arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare merchant webhook queue: %w", err)
	}

	// Bind queue to topic
	err = ch.QueueBind(
		q.Name,
		"webhook.merchant.#",
		queue.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind merchant webhook queue: %w", err)
	}

	// Set prefetch QoS
	if err := ch.Qos(10, 0, false); err != nil {
		log.Printf("Warning: failed to set Qos on webhook channel: %v\n", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer tag
		false, // autoAck = false (manual ack on 2xx response)
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start merchant webhook consumer: %w", err)
	}

	fmt.Println("[RabbitMQ Worker] Subscribed to durable 'merchant_webhooks_queue' (DLQ Enabled)")

	go func() {
		for d := range msgs {
			var event dto.MerchantWebhookEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("[RabbitMQ Webhook] Failed to deserialize payload: %v\n", err)
				d.Nack(false, false) // unroutable message sent to DLQ
				continue
			}

			// Read current retry count from headers
			var retryCount int32 = 0
			if val, ok := d.Headers["x-retry-count"]; ok {
				if rc, ok := val.(int32); ok {
					retryCount = rc
				} else if rci, ok := val.(int); ok {
					retryCount = int32(rci)
				}
			}

			attempt := retryCount + 1
			err := sendMerchantWebhook(&event, attempt)
			if err == nil {
				// 2xx Success -> Acknowledge message
				d.Ack(false)
			} else {
				// Failed delivery
				if retryCount < 3 {
					backoffDelay := time.Duration(retryCount+1) * 2 * time.Second
					log.Printf("[RabbitMQ Webhook] Retry %d/3 for order %s scheduled in %v (Error: %v)\n",
						attempt, event.OrderID, backoffDelay, err)

					time.Sleep(backoffDelay)

					// Republish with incremented retry count
					if globalRabbitMQClient != nil {
						ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
						_ = globalRabbitMQClient.PublishMerchantWebhook(ctx, "webhook.merchant.event", &event, retryCount+1)
						cancel()
					}
					d.Ack(false)
				} else {
					// Max retries exceeded -> Nack without requeue to trigger RabbitMQ DLX -> merchant_webhooks_dlq
					log.Printf("[RabbitMQ Webhook DLQ] Max retries (3/3) exceeded for event %s (Order: %s). Diverted to Dead Letter Queue.\n",
						event.Event, event.OrderID)
					d.Nack(false, false)
				}
			}
		}
	}()

	return nil
}

func sendWebhook(tx *dto.TransactionWebhookEvent) {
	webhookUrl := os.Getenv("DEFAULT_WEBHOOK_URL")
	if webhookUrl == "" {
		// No default webhook endpoint configured
		return
	}
	jsonData, _ := json.Marshal(tx)

	resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[Webhook error] Process ID %s couldn't sent %v\n", tx.TransactionID, err)
		return
	}

	defer resp.Body.Close()
	fmt.Printf("[Webhook success] Process ID %s sent!\n", tx.TransactionID)
}

func sendMerchantWebhook(event *dto.MerchantWebhookEvent, attempt int32) error {
	if event.WebhookURL == "" {
		return nil
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", event.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TyrFintech-Webhook-Dispatcher/2.0 (RabbitMQ-Durable)")
	req.Header.Set("X-TyrFintech-Event", event.Event)
	req.Header.Set("X-TyrFintech-Delivery-Attempt", fmt.Sprintf("%d", attempt))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network delivery error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("remote endpoint returned HTTP %d", resp.StatusCode)
	}

	log.Printf("[Merchant Webhook Success] Event '%s' delivered to %s (Status: %d, Attempt: %d)\n",
		event.Event, event.WebhookURL, resp.StatusCode, attempt)
	return nil
}
