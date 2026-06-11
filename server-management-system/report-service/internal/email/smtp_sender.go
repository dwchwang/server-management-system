package email

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/vcs-sms/report-service/config"
	"gopkg.in/gomail.v2"
)

// EmailSender defines the interface for sending HTML emails.
type EmailSender interface {
	SendHTML(ctx context.Context, to string, subject string, htmlBody string) error
}

// GmailSender implements EmailSender using Gmail SMTP.
type GmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

type smtpDialer interface {
	DialAndSend(...*gomail.Message) error
}

var newSMTPDialerOverride func(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer

func createSMTPDialer(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer {
	if newSMTPDialerOverride != nil {
		return newSMTPDialerOverride(host, port, username, password, tlsConfig)
	}

	d := gomail.NewDialer(host, port, username, password)
	d.TLSConfig = tlsConfig
	return d
}

// NewGmailSender creates a new GmailSender from SMTP config.
func NewGmailSender(cfg config.SMTPConfig) *GmailSender {
	return &GmailSender{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
	}
}

// SendHTML sends an HTML email via SMTP.
func (s *GmailSender) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	tlsConfig := &tls.Config{
		ServerName: s.host,
		MinVersion: tls.VersionTLS12,
	}
	d := createSMTPDialer(s.host, s.port, s.username, s.password, tlsConfig)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
