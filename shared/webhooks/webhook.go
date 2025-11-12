package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWebhookNotFound     = errors.New("webhook not found")
	ErrInvalidSignature    = errors.New("invalid webhook signature")
	ErrDeliveryFailed      = errors.New("webhook delivery failed")
	ErrMaxRetriesExceeded  = errors.New("max retries exceeded")
)

// EventType defines the type of webhook event
type EventType string

const (
	EventUserCreated       EventType = "user.created"
	EventUserUpdated       EventType = "user.updated"
	EventUserDeleted       EventType = "user.deleted"
	EventPaymentSucceeded  EventType = "payment.succeeded"
	EventPaymentFailed     EventType = "payment.failed"
	EventSubscriptionCreated EventType = "subscription.created"
	EventSubscriptionUpdated EventType = "subscription.updated"
	EventSubscriptionCanceled EventType = "subscription.canceled"
	EventCustom            EventType = "custom"
)

// Webhook represents a webhook endpoint configuration
type Webhook struct {
	ID          uuid.UUID   `json:"id"`
	TenantID    uuid.UUID   `json:"tenant_id"`
	URL         string      `json:"url"`
	Secret      string      `json:"secret"` // For HMAC signature
	Events      []EventType `json:"events"` // Events to subscribe to
	Active      bool        `json:"active"`
	Description string      `json:"description,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"` // Custom headers
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// Event represents a webhook event
type Event struct {
	ID        uuid.UUID              `json:"id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	Type      EventType              `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// Delivery represents a webhook delivery attempt
type Delivery struct {
	ID           uuid.UUID  `json:"id"`
	WebhookID    uuid.UUID  `json:"webhook_id"`
	EventID      uuid.UUID  `json:"event_id"`
	Status       string     `json:"status"` // pending, success, failed
	StatusCode   int        `json:"status_code,omitempty"`
	Response     string     `json:"response,omitempty"`
	Error        string     `json:"error,omitempty"`
	Attempt      int        `json:"attempt"`
	NextRetryAt  *time.Time `json:"next_retry_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
}

// Repository handles webhook data access
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new webhook repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create creates a new webhook
func (r *Repository) Create(ctx context.Context, webhook *Webhook) error {
	query := `
		INSERT INTO webhooks (id, tenant_id, url, secret, events, active, description, headers, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)

	_, err := r.db.ExecContext(ctx, query,
		webhook.ID, webhook.TenantID, webhook.URL, webhook.Secret,
		eventsJSON, webhook.Active, webhook.Description, headersJSON,
		webhook.CreatedAt, webhook.UpdatedAt,
	)

	return err
}

// GetByID retrieves a webhook by ID
func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*Webhook, error) {
	query := `
		SELECT id, tenant_id, url, secret, events, active, description, headers, created_at, updated_at
		FROM webhooks
		WHERE id = $1
	`

	var webhook Webhook
	var eventsJSON, headersJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&webhook.ID, &webhook.TenantID, &webhook.URL, &webhook.Secret,
		&eventsJSON, &webhook.Active, &webhook.Description, &headersJSON,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrWebhookNotFound
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(eventsJSON, &webhook.Events)
	json.Unmarshal(headersJSON, &webhook.Headers)

	return &webhook, nil
}

// ListByTenant retrieves all webhooks for a tenant
func (r *Repository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*Webhook, error) {
	query := `
		SELECT id, tenant_id, url, secret, events, active, description, headers, created_at, updated_at
		FROM webhooks
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []*Webhook
	for rows.Next() {
		var webhook Webhook
		var eventsJSON, headersJSON []byte

		err := rows.Scan(
			&webhook.ID, &webhook.TenantID, &webhook.URL, &webhook.Secret,
			&eventsJSON, &webhook.Active, &webhook.Description, &headersJSON,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(eventsJSON, &webhook.Events)
		json.Unmarshal(headersJSON, &webhook.Headers)

		webhooks = append(webhooks, &webhook)
	}

	return webhooks, rows.Err()
}

// Update updates a webhook
func (r *Repository) Update(ctx context.Context, webhook *Webhook) error {
	query := `
		UPDATE webhooks
		SET url = $2, events = $3, active = $4, description = $5, headers = $6, updated_at = $7
		WHERE id = $1
	`

	eventsJSON, _ := json.Marshal(webhook.Events)
	headersJSON, _ := json.Marshal(webhook.Headers)

	result, err := r.db.ExecContext(ctx, query,
		webhook.ID, webhook.URL, eventsJSON, webhook.Active,
		webhook.Description, headersJSON, webhook.UpdatedAt,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// Delete deletes a webhook
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrWebhookNotFound
	}

	return nil
}

// SaveDelivery saves a delivery attempt
func (r *Repository) SaveDelivery(ctx context.Context, delivery *Delivery) error {
	query := `
		INSERT INTO webhook_deliveries
		(id, webhook_id, event_id, status, status_code, response, error, attempt, next_retry_at, created_at, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		delivery.ID, delivery.WebhookID, delivery.EventID,
		delivery.Status, delivery.StatusCode, delivery.Response,
		delivery.Error, delivery.Attempt, delivery.NextRetryAt,
		delivery.CreatedAt, delivery.DeliveredAt,
	)

	return err
}

// GetPendingDeliveries retrieves deliveries that need to be retried
func (r *Repository) GetPendingDeliveries(ctx context.Context) ([]*Delivery, error) {
	query := `
		SELECT id, webhook_id, event_id, status, status_code, response, error, attempt, next_retry_at, created_at, delivered_at
		FROM webhook_deliveries
		WHERE status = 'pending' AND next_retry_at <= $1
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []*Delivery
	for rows.Next() {
		var delivery Delivery
		err := rows.Scan(
			&delivery.ID, &delivery.WebhookID, &delivery.EventID,
			&delivery.Status, &delivery.StatusCode, &delivery.Response,
			&delivery.Error, &delivery.Attempt, &delivery.NextRetryAt,
			&delivery.CreatedAt, &delivery.DeliveredAt,
		)
		if err != nil {
			return nil, err
		}

		deliveries = append(deliveries, &delivery)
	}

	return deliveries, rows.Err()
}

// Service handles webhook business logic
type Service struct {
	repo       *Repository
	httpClient *http.Client
	maxRetries int
}

// NewService creates a new webhook service
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		maxRetries: 5,
	}
}

// Dispatch dispatches an event to all subscribed webhooks
func (s *Service) Dispatch(ctx context.Context, event *Event) error {
	// Get all active webhooks for this tenant
	webhooks, err := s.repo.ListByTenant(ctx, event.TenantID)
	if err != nil {
		return err
	}

	// Filter webhooks that are subscribed to this event type
	for _, webhook := range webhooks {
		if !webhook.Active {
			continue
		}

		// Check if webhook is subscribed to this event type
		subscribed := false
		for _, eventType := range webhook.Events {
			if eventType == event.Type || eventType == "*" {
				subscribed = true
				break
			}
		}

		if !subscribed {
			continue
		}

		// Create delivery
		delivery := &Delivery{
			ID:        uuid.New(),
			WebhookID: webhook.ID,
			EventID:   event.ID,
			Status:    "pending",
			Attempt:   0,
			CreatedAt: time.Now(),
		}

		// Attempt delivery
		go s.deliver(ctx, webhook, event, delivery)
	}

	return nil
}

// deliver attempts to deliver a webhook
func (s *Service) deliver(ctx context.Context, webhook *Webhook, event *Event, delivery *Delivery) error {
	delivery.Attempt++

	// Prepare payload
	payload := map[string]interface{}{
		"id":         event.ID,
		"type":       event.Type,
		"data":       event.Data,
		"created_at": event.Created,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, bytes.NewReader(payloadJSON))
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Marimo-Webhook/1.0")
	req.Header.Set("X-Webhook-ID", webhook.ID.String())
	req.Header.Set("X-Event-ID", event.ID.String())
	req.Header.Set("X-Event-Type", string(event.Type))

	// Add custom headers
	for key, value := range webhook.Headers {
		req.Header.Set(key, value)
	}

	// Add HMAC signature
	signature := s.generateSignature(payloadJSON, webhook.Secret)
	req.Header.Set("X-Webhook-Signature", signature)

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		delivery.Status = "failed"
		delivery.Error = err.Error()
		s.scheduleRetry(delivery)
		s.repo.SaveDelivery(ctx, delivery)
		return err
	}
	defer resp.Body.Close()

	// Read response
	responseBody, _ := io.ReadAll(resp.Body)
	delivery.StatusCode = resp.StatusCode
	delivery.Response = string(responseBody)

	// Check if successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		delivery.Status = "success"
		now := time.Now()
		delivery.DeliveredAt = &now
	} else {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		s.scheduleRetry(delivery)
	}

	return s.repo.SaveDelivery(ctx, delivery)
}

// scheduleRetry schedules a retry with exponential backoff
func (s *Service) scheduleRetry(delivery *Delivery) {
	if delivery.Attempt >= s.maxRetries {
		delivery.Status = "failed"
		delivery.Error = fmt.Sprintf("%s: %s", ErrMaxRetriesExceeded.Error(), delivery.Error)
		delivery.NextRetryAt = nil
		return
	}

	// Exponential backoff: 1m, 5m, 15m, 1h, 6h
	delays := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		6 * time.Hour,
	}

	delay := delays[min(delivery.Attempt-1, len(delays)-1)]
	nextRetry := time.Now().Add(delay)
	delivery.NextRetryAt = &nextRetry
	delivery.Status = "pending"
}

// generateSignature generates HMAC-SHA256 signature
func (s *Service) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies webhook signature
func VerifySignature(payload []byte, signature, secret string) bool {
	expected := generateSignatureStatic(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

func generateSignatureStatic(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
