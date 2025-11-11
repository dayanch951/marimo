package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageQueue handles RabbitMQ operations
type MessageQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Message represents a message in the queue
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry"`
}

// NewMessageQueue creates a new RabbitMQ client
func NewMessageQueue(url string) (*MessageQueue, error) {
	if url == "" {
		url = os.Getenv("RABBITMQ_URL")
		if url == "" {
			url = "amqp://guest:guest@localhost:5672/"
		}
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	log.Println("Connected to RabbitMQ")

	return &MessageQueue{
		conn:    conn,
		channel: channel,
	}, nil
}

// DeclareQueue declares a queue with durability
func (mq *MessageQueue) DeclareQueue(queueName string) error {
	_, err := mq.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Printf("Queue %s declared", queueName)
	return nil
}

// DeclareExchange declares an exchange
func (mq *MessageQueue) DeclareExchange(exchangeName, exchangeType string) error {
	err := mq.channel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type (fanout, direct, topic, headers)
		true,         // durable
		false,        // auto-delete
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("Exchange %s declared (type: %s)", exchangeName, exchangeType)
	return nil
}

// BindQueue binds a queue to an exchange with a routing key
func (mq *MessageQueue) BindQueue(queueName, exchangeName, routingKey string) error {
	err := mq.channel.QueueBind(
		queueName,    // queue name
		routingKey,   // routing key
		exchangeName, // exchange
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	log.Printf("Queue %s bound to exchange %s with routing key %s", queueName, exchangeName, routingKey)
	return nil
}

// Publish sends a message to a queue
func (mq *MessageQueue) Publish(queueName string, message Message) error {
	message.Timestamp = time.Now()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = mq.channel.Publish(
		"",        // exchange
		queueName, // routing key (queue name for direct send)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    message.Timestamp,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// PublishToExchange sends a message to an exchange with a routing key
func (mq *MessageQueue) PublishToExchange(exchangeName, routingKey string, message Message) error {
	message.Timestamp = time.Now()

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = mq.channel.Publish(
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    message.Timestamp,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message to exchange: %w", err)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (mq *MessageQueue) Consume(queueName string, handler func(Message) error) error {
	// Set QoS to process one message at a time
	err := mq.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := mq.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack (manual ack for reliability)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	log.Printf("Started consuming from queue %s", queueName)

	// Process messages
	go func() {
		for msg := range msgs {
			var message Message
			err := json.Unmarshal(msg.Body, &message)
			if err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				msg.Nack(false, false) // Don't requeue malformed messages
				continue
			}

			// Handle the message
			err = handler(message)
			if err != nil {
				log.Printf("Failed to handle message: %v", err)

				// Retry logic - requeue up to 3 times
				if message.Retry < 3 {
					message.Retry++
					log.Printf("Requeuing message (retry %d/3)", message.Retry)
					msg.Nack(false, true) // Requeue
				} else {
					log.Printf("Max retries reached, sending to dead letter queue")
					msg.Nack(false, false) // Don't requeue
					// In production, send to dead letter queue
				}
			} else {
				msg.Ack(false) // Acknowledge successful processing
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (mq *MessageQueue) Close() error {
	if mq.channel != nil {
		if err := mq.channel.Close(); err != nil {
			return fmt.Errorf("failed to close channel: %w", err)
		}
	}

	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	log.Println("RabbitMQ connection closed")
	return nil
}

// GetQueueInfo returns information about a queue
func (mq *MessageQueue) GetQueueInfo(queueName string) (int, int, error) {
	queue, err := mq.channel.QueueInspect(queueName)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return queue.Messages, queue.Consumers, nil
}

// PurgeQueue removes all messages from a queue
func (mq *MessageQueue) PurgeQueue(queueName string) error {
	_, err := mq.channel.QueuePurge(queueName, false)
	if err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}

	log.Printf("Queue %s purged", queueName)
	return nil
}

// DeleteQueue deletes a queue
func (mq *MessageQueue) DeleteQueue(queueName string) error {
	_, err := mq.channel.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}

	log.Printf("Queue %s deleted", queueName)
	return nil
}
