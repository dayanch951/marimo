package async

import (
	"fmt"
	"log"

	"github.com/dayanch951/marimo/shared/queue"
	"github.com/google/uuid"
)

// EventType represents different types of events
type EventType string

const (
	EventUserRegistered EventType = "user.registered"
	EventUserLogin      EventType = "user.login"
	EventUserLogout     EventType = "user.logout"
	EventAuditLog       EventType = "audit.log"
	EventEmailSend      EventType = "email.send"
	EventNotification   EventType = "notification.send"
)

// Queue names for different event types
const (
	QueueEmail        = "email_queue"
	QueueNotification = "notification_queue"
	QueueAudit        = "audit_queue"
	QueueEvents       = "events_queue"
)

// EventPublisher publishes events to RabbitMQ
type EventPublisher struct {
	mq *queue.MessageQueue
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(rabbitmqURL string) (*EventPublisher, error) {
	mq, err := queue.NewMessageQueue(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create message queue: %w", err)
	}

	// Declare queues
	queues := []string{
		QueueEmail,
		QueueNotification,
		QueueAudit,
		QueueEvents,
	}

	for _, q := range queues {
		if err := mq.DeclareQueue(q); err != nil {
			return nil, fmt.Errorf("failed to declare queue %s: %w", q, err)
		}
	}

	return &EventPublisher{mq: mq}, nil
}

// PublishUserRegistered publishes a user registration event
func (ep *EventPublisher) PublishUserRegistered(userID, email string) error {
	msg := queue.Message{
		ID:   uuid.New().String(),
		Type: string(EventUserRegistered),
		Payload: map[string]interface{}{
			"user_id": userID,
			"email":   email,
		},
	}

	// Send to multiple queues
	if err := ep.mq.Publish(QueueEmail, msg); err != nil {
		return fmt.Errorf("failed to publish to email queue: %w", err)
	}

	if err := ep.mq.Publish(QueueAudit, msg); err != nil {
		return fmt.Errorf("failed to publish to audit queue: %w", err)
	}

	if err := ep.mq.Publish(QueueEvents, msg); err != nil {
		return fmt.Errorf("failed to publish to events queue: %w", err)
	}

	log.Printf("Published user registration event: %s", userID)
	return nil
}

// PublishUserLogin publishes a user login event
func (ep *EventPublisher) PublishUserLogin(userID, email, ipAddress string) error {
	msg := queue.Message{
		ID:   uuid.New().String(),
		Type: string(EventUserLogin),
		Payload: map[string]interface{}{
			"user_id":    userID,
			"email":      email,
			"ip_address": ipAddress,
		},
	}

	if err := ep.mq.Publish(QueueAudit, msg); err != nil {
		return fmt.Errorf("failed to publish login event: %w", err)
	}

	log.Printf("Published user login event: %s", userID)
	return nil
}

// PublishEmail publishes an email send event
func (ep *EventPublisher) PublishEmail(to, subject, body string) error {
	msg := queue.Message{
		ID:   uuid.New().String(),
		Type: string(EventEmailSend),
		Payload: map[string]interface{}{
			"to":      to,
			"subject": subject,
			"body":    body,
		},
	}

	if err := ep.mq.Publish(QueueEmail, msg); err != nil {
		return fmt.Errorf("failed to publish email event: %w", err)
	}

	log.Printf("Published email event to: %s", to)
	return nil
}

// PublishAuditLog publishes an audit log event
func (ep *EventPublisher) PublishAuditLog(userID, action, resource string, metadata map[string]interface{}) error {
	payload := map[string]interface{}{
		"user_id":  userID,
		"action":   action,
		"resource": resource,
	}

	// Merge metadata
	for k, v := range metadata {
		payload[k] = v
	}

	msg := queue.Message{
		ID:      uuid.New().String(),
		Type:    string(EventAuditLog),
		Payload: payload,
	}

	if err := ep.mq.Publish(QueueAudit, msg); err != nil {
		return fmt.Errorf("failed to publish audit log: %w", err)
	}

	log.Printf("Published audit log: %s - %s", action, resource)
	return nil
}

// Close closes the event publisher
func (ep *EventPublisher) Close() error {
	return ep.mq.Close()
}

// EventHandler handles different types of events
type EventHandler struct {
	mq *queue.MessageQueue
}

// NewEventHandler creates a new event handler
func NewEventHandler(rabbitmqURL string) (*EventHandler, error) {
	mq, err := queue.NewMessageQueue(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create message queue: %w", err)
	}

	return &EventHandler{mq: mq}, nil
}

// StartEmailWorker starts consuming email events
func (eh *EventHandler) StartEmailWorker() error {
	log.Println("Starting email worker...")

	handler := func(msg queue.Message) error {
		log.Printf("Processing email: %s", msg.Type)

		to, _ := msg.Payload["to"].(string)
		subject, _ := msg.Payload["subject"].(string)
		body, _ := msg.Payload["body"].(string)

		// Simulate sending email
		log.Printf("Sending email to %s: %s", to, subject)

		// In production, use actual email service (SendGrid, AWS SES, etc.)
		// For now, just log
		log.Printf("Email sent successfully: %s", msg.ID)

		return nil
	}

	return eh.mq.Consume(QueueEmail, handler)
}

// StartAuditWorker starts consuming audit log events
func (eh *EventHandler) StartAuditWorker() error {
	log.Println("Starting audit worker...")

	handler := func(msg queue.Message) error {
		log.Printf("Processing audit log: %s", msg.Type)

		userID, _ := msg.Payload["user_id"].(string)
		action, _ := msg.Payload["action"].(string)
		resource, _ := msg.Payload["resource"].(string)

		// In production, write to database or log aggregation service
		log.Printf("Audit: User %s performed %s on %s", userID, action, resource)

		return nil
	}

	return eh.mq.Consume(QueueAudit, handler)
}

// StartNotificationWorker starts consuming notification events
func (eh *EventHandler) StartNotificationWorker() error {
	log.Println("Starting notification worker...")

	handler := func(msg queue.Message) error {
		log.Printf("Processing notification: %s", msg.Type)

		// In production, send push notifications, SMS, etc.
		log.Printf("Notification sent: %s", msg.ID)

		return nil
	}

	return eh.mq.Consume(QueueNotification, handler)
}

// Close closes the event handler
func (eh *EventHandler) Close() error {
	return eh.mq.Close()
}
