package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailService handles email sending
type EmailService struct {
	config EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService() *EmailService {
	return &EmailService{
		config: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnv("SMTP_PORT", "587"),
			SMTPUsername: getEnv("SMTP_USERNAME", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "noreply@marimo.dev"),
			FromName:     getEnv("FROM_NAME", "Marimo ERP"),
		},
	}
}

// EmailMessage represents an email to send
type EmailMessage struct {
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Content  []byte
}

// SendEmail sends an email using SMTP
func (es *EmailService) SendEmail(msg EmailMessage) error {
	// Build email headers
	from := fmt.Sprintf("%s <%s>", es.config.FromName, es.config.FromEmail)

	// Create message
	var body bytes.Buffer
	body.WriteString(fmt.Sprintf("From: %s\r\n", from))
	body.WriteString(fmt.Sprintf("To: %s\r\n", msg.To[0]))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	body.WriteString("MIME-Version: 1.0\r\n")

	if msg.HTMLBody != "" {
		body.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		body.WriteString("\r\n")
		body.WriteString(msg.HTMLBody)
	} else {
		body.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		body.WriteString("\r\n")
		body.WriteString(msg.Body)
	}

	// SMTP authentication
	auth := smtp.PlainAuth("", es.config.SMTPUsername, es.config.SMTPPassword, es.config.SMTPHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", es.config.SMTPHost, es.config.SMTPPort)
	err := smtp.SendMail(addr, auth, es.config.FromEmail, msg.To, body.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email sent successfully to %v", msg.To)
	return nil
}

// SendWelcomeEmail sends a welcome email to new users
func (es *EmailService) SendWelcomeEmail(to, name string) error {
	tmpl := template.Must(template.New("welcome").Parse(welcomeTemplate))

	var body bytes.Buffer
	err := tmpl.Execute(&body, map[string]string{
		"Name": name,
	})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return es.SendEmail(EmailMessage{
		To:       []string{to},
		Subject:  "Welcome to Marimo ERP!",
		HTMLBody: body.String(),
	})
}

// SendPasswordResetEmail sends a password reset email
func (es *EmailService) SendPasswordResetEmail(to, resetLink string) error {
	tmpl := template.Must(template.New("reset").Parse(passwordResetTemplate))

	var body bytes.Buffer
	err := tmpl.Execute(&body, map[string]string{
		"ResetLink": resetLink,
	})
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return es.SendEmail(EmailMessage{
		To:       []string{to},
		Subject:  "Password Reset Request",
		HTMLBody: body.String(),
	})
}

// SendNotificationEmail sends a generic notification email
func (es *EmailService) SendNotificationEmail(to, subject, message string) error {
	return es.SendEmail(EmailMessage{
		To:      []string{to},
		Subject: subject,
		Body:    message,
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Email templates
const welcomeTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #667eea; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to Marimo ERP!</h1>
        </div>
        <div class="content">
            <h2>Hello {{.Name}}!</h2>
            <p>Thank you for joining Marimo ERP. We're excited to have you on board!</p>
            <p>With Marimo ERP, you can:</p>
            <ul>
                <li>Manage your business operations efficiently</li>
                <li>Track inventory and sales</li>
                <li>Generate reports and analytics</li>
                <li>Collaborate with your team</li>
            </ul>
            <p>Get started by logging into your account:</p>
            <a href="https://marimo.dev/login" class="button">Login to Your Account</a>
        </div>
        <div class="footer">
            <p>© 2024 Marimo ERP. All rights reserved.</p>
            <p>If you have any questions, please contact us at support@marimo.dev</p>
        </div>
    </div>
</body>
</html>
`

const passwordResetTemplate = `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #f44336; color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .button { display: inline-block; padding: 12px 30px; background: #f44336; color: white; text-decoration: none; border-radius: 5px; margin-top: 20px; }
        .warning { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 30px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>We received a request to reset your password for your Marimo ERP account.</p>
            <p>Click the button below to reset your password:</p>
            <a href="{{.ResetLink}}" class="button">Reset Password</a>
            <div class="warning">
                <strong>Security Notice:</strong> This link will expire in 1 hour. If you didn't request a password reset, please ignore this email or contact support if you have concerns.
            </div>
            <p>If the button doesn't work, copy and paste this link into your browser:</p>
            <p style="word-break: break-all; color: #667eea;">{{.ResetLink}}</p>
        </div>
        <div class="footer">
            <p>© 2024 Marimo ERP. All rights reserved.</p>
            <p>Never share this email with anyone for security reasons.</p>
        </div>
    </div>
</body>
</html>
`
