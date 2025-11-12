package integrations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSendGridNotConfigured = errors.New("SendGrid is not configured")
	ErrInvalidEmail          = errors.New("invalid email address")
	ErrEmailSendFailed       = errors.New("failed to send email")
)

// SendGridConfig holds SendGrid API configuration
type SendGridConfig struct {
	APIKey          string
	FromEmail       string
	FromName        string
	ReplyTo         string
	UnsubscribeURL  string
}

// SendGridClient wraps SendGrid API operations
type SendGridClient struct {
	config SendGridConfig
}

// NewSendGridClient creates a new SendGrid client
func NewSendGridClient(config SendGridConfig) *SendGridClient {
	return &SendGridClient{config: config}
}

// EmailAddress represents an email address with name
type EmailAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// EmailAttachment represents an email attachment
type EmailAttachment struct {
	Filename    string `json:"filename"`
	Content     []byte `json:"content"`
	Type        string `json:"type"`
	Disposition string `json:"disposition"` // "attachment" or "inline"
}

// EmailMessage represents an email to be sent
type EmailMessage struct {
	To          []EmailAddress    `json:"to"`
	CC          []EmailAddress    `json:"cc,omitempty"`
	BCC         []EmailAddress    `json:"bcc,omitempty"`
	Subject     string            `json:"subject"`
	TextContent string            `json:"text_content,omitempty"`
	HTMLContent string            `json:"html_content,omitempty"`
	Attachments []EmailAttachment `json:"attachments,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Categories  []string          `json:"categories,omitempty"`
	CustomArgs  map[string]string `json:"custom_args,omitempty"`
	SendAt      *time.Time        `json:"send_at,omitempty"` // Scheduled send
	TrackingSettings TrackingSettings `json:"tracking_settings,omitempty"`
}

// TrackingSettings controls email tracking
type TrackingSettings struct {
	ClickTracking       bool `json:"click_tracking"`
	OpenTracking        bool `json:"open_tracking"`
	SubscriptionTracking bool `json:"subscription_tracking"`
}

// EmailResponse represents the response after sending an email
type EmailResponse struct {
	MessageID string    `json:"message_id"`
	Status    string    `json:"status"`
	SentAt    time.Time `json:"sent_at"`
}

// SendEmail sends a single email
func (sg *SendGridClient) SendEmail(ctx context.Context, message *EmailMessage) (*EmailResponse, error) {
	if len(message.To) == 0 {
		return nil, ErrInvalidEmail
	}

	// In real implementation, use SendGrid SDK:
	// from := mail.NewEmail(sg.config.FromName, sg.config.FromEmail)
	// to := mail.NewEmail(message.To[0].Name, message.To[0].Email)
	// msg := mail.NewSingleEmail(from, message.Subject, to, message.TextContent, message.HTMLContent)
	// client := sendgrid.NewSendClient(sg.config.APIKey)
	// response, err := client.Send(msg)

	// Mock implementation
	response := &EmailResponse{
		MessageID: fmt.Sprintf("msg_%s", uuid.New().String()[:8]),
		Status:    "sent",
		SentAt:    time.Now(),
	}

	return response, nil
}

// SendBulkEmail sends emails to multiple recipients
func (sg *SendGridClient) SendBulkEmail(ctx context.Context, messages []*EmailMessage) ([]*EmailResponse, error) {
	var responses []*EmailResponse

	for _, message := range messages {
		response, err := sg.SendEmail(ctx, message)
		if err != nil {
			return nil, fmt.Errorf("failed to send email to %s: %w", message.To[0].Email, err)
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// Template represents a SendGrid template
type Template struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Versions []TemplateVersion `json:"versions"`
}

// TemplateVersion represents a version of a template
type TemplateVersion struct {
	ID          string    `json:"id"`
	Active      bool      `json:"active"`
	Name        string    `json:"name"`
	Subject     string    `json:"subject"`
	HTMLContent string    `json:"html_content"`
	TextContent string    `json:"plain_content"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SendTemplateEmail sends an email using a template
func (sg *SendGridClient) SendTemplateEmail(ctx context.Context, templateID string, to []EmailAddress, dynamicData map[string]interface{}) (*EmailResponse, error) {
	// In real implementation:
	// from := mail.NewEmail(sg.config.FromName, sg.config.FromEmail)
	// msg := mail.NewV3Mail()
	// msg.SetFrom(from)
	// msg.SetTemplateID(templateID)
	// personalization := mail.NewPersonalization()
	// for _, recipient := range to {
	//     personalization.AddTos(mail.NewEmail(recipient.Name, recipient.Email))
	// }
	// for key, value := range dynamicData {
	//     personalization.SetDynamicTemplateData(key, value)
	// }
	// msg.AddPersonalizations(personalization)

	// Mock implementation
	response := &EmailResponse{
		MessageID: fmt.Sprintf("msg_%s", uuid.New().String()[:8]),
		Status:    "sent",
		SentAt:    time.Now(),
	}

	return response, nil
}

// ContactListCreateParams parameters for creating a contact list
type ContactListCreateParams struct {
	Name string `json:"name"`
}

// ContactList represents a SendGrid contact list
type ContactList struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	ContactCount int       `json:"contact_count"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateContactList creates a new contact list
func (sg *SendGridClient) CreateContactList(ctx context.Context, params ContactListCreateParams) (*ContactList, error) {
	// Mock implementation
	list := &ContactList{
		ID:           fmt.Sprintf("list_%s", uuid.New().String()[:8]),
		Name:         params.Name,
		ContactCount: 0,
		CreatedAt:    time.Now(),
	}

	return list, nil
}

// Contact represents a SendGrid contact
type Contact struct {
	ID         string            `json:"id"`
	Email      string            `json:"email"`
	FirstName  string            `json:"first_name,omitempty"`
	LastName   string            `json:"last_name,omitempty"`
	CustomFields map[string]string `json:"custom_fields,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// AddContactToList adds a contact to a list
func (sg *SendGridClient) AddContactToList(ctx context.Context, listID string, contact *Contact) error {
	// In real implementation:
	// Use SendGrid Marketing Campaigns API

	// Mock implementation
	return nil
}

// EmailStats represents email statistics
type EmailStats struct {
	Date       time.Time `json:"date"`
	Delivered  int       `json:"delivered"`
	Opens      int       `json:"opens"`
	UniqueOpens int      `json:"unique_opens"`
	Clicks     int       `json:"clicks"`
	UniqueClicks int     `json:"unique_clicks"`
	Bounces    int       `json:"bounces"`
	Spam       int       `json:"spam"`
	Unsubscribes int     `json:"unsubscribes"`
}

// GetStats retrieves email statistics for a date range
func (sg *SendGridClient) GetStats(ctx context.Context, startDate, endDate time.Time) ([]EmailStats, error) {
	// In real implementation:
	// Use SendGrid Stats API

	// Mock implementation
	stats := []EmailStats{
		{
			Date:         startDate,
			Delivered:    1000,
			Opens:        500,
			UniqueOpens:  400,
			Clicks:       200,
			UniqueClicks: 150,
			Bounces:      10,
			Spam:         5,
			Unsubscribes: 3,
		},
	}

	return stats, nil
}

// UnsubscribeGroup represents a suppression group
type UnsubscribeGroup struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateUnsubscribeGroup creates a new unsubscribe group
func (sg *SendGridClient) CreateUnsubscribeGroup(ctx context.Context, name, description string) (*UnsubscribeGroup, error) {
	// Mock implementation
	group := &UnsubscribeGroup{
		ID:          int(time.Now().Unix()),
		Name:        name,
		Description: description,
	}

	return group, nil
}

// WebhookEvent represents a SendGrid webhook event
type WebhookEvent struct {
	Email     string            `json:"email"`
	Event     string            `json:"event"` // delivered, open, click, bounce, etc.
	Timestamp int64             `json:"timestamp"`
	Category  []string          `json:"category,omitempty"`
	CustomArgs map[string]string `json:"custom_args,omitempty"`
}

// HandleWebhook processes SendGrid webhook events
func (sg *SendGridClient) HandleWebhook(ctx context.Context, events []WebhookEvent) error {
	for _, event := range events {
		switch event.Event {
		case "delivered":
			// Handle delivered event
		case "open":
			// Handle open event
		case "click":
			// Handle click event
		case "bounce":
			// Handle bounce event
		case "dropped":
			// Handle dropped event
		case "spam_report":
			// Handle spam report
		case "unsubscribe":
			// Handle unsubscribe
		}
	}

	return nil
}

// EmailCampaign represents a marketing campaign
type EmailCampaign struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Subject     string     `json:"subject"`
	SenderID    int        `json:"sender_id"`
	ListIDs     []string   `json:"list_ids"`
	Status      string     `json:"status"` // draft, scheduled, sent
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// CreateCampaign creates a new email campaign
func (sg *SendGridClient) CreateCampaign(ctx context.Context, campaign *EmailCampaign) (*EmailCampaign, error) {
	// Mock implementation
	campaign.ID = fmt.Sprintf("campaign_%s", uuid.New().String()[:8])
	campaign.Status = "draft"
	campaign.CreatedAt = time.Now()

	return campaign, nil
}

// SendCampaign sends or schedules a campaign
func (sg *SendGridClient) SendCampaign(ctx context.Context, campaignID string, scheduleAt *time.Time) error {
	// In real implementation:
	// Use SendGrid Marketing Campaigns API

	// Mock implementation
	return nil
}
