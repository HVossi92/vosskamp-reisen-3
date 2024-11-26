package services

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/joho/godotenv"
)

type EmailConfig struct {
	Host     string
	Port     string
	From     string
	Password string
}

type EmailService struct {
	config EmailConfig
}

func NewEmailService() *EmailService {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	// Get email configuration from environment variables
	config := EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		From:     os.Getenv("EMAIL"),
		Password: os.Getenv("EMAIL_PASSWORD"),
	}

	// Validate configuration
	if config.Host == "" || config.Port == "" || config.From == "" || config.Password == "" {
		panic("missing required email configuration")
	}

	return &EmailService{config: config}
}

func (s *EmailService) SendEmail(to []string, subject, body string) error {
	// Compose email
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("Subject: %s\n%s\n%s", subject, mime, body))

	// Create authentication
	auth := smtp.PlainAuth("", s.config.From, s.config.Password, s.config.Host)

	// Create SMTP address
	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Send email
	if err := smtp.SendMail(addr, auth, s.config.From, to, msg); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// Helper function to send a template-based email
func (s *EmailService) SendTemplatedEmail(to []string, subject string, templateData interface{}, template func(interface{}) string) error {
	// Generate email body from template
	body := template(templateData)

	// Send email using the generated body
	return s.SendEmail(to, subject, body)
}
