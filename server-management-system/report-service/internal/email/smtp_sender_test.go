package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vcs-sms/report-service/config"
	"gopkg.in/gomail.v2"
)

type fakeSMTPDialer struct {
	sendFunc func(...*gomail.Message) error
}

func (d *fakeSMTPDialer) DialAndSend(messages ...*gomail.Message) error {
	if d.sendFunc != nil {
		return d.sendFunc(messages...)
	}
	return nil
}

func withFakeDialer(t *testing.T, factory func(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer) {
	t.Helper()
	original := newSMTPDialerOverride
	newSMTPDialerOverride = factory
	t.Cleanup(func() {
		newSMTPDialerOverride = original
	})
}

func TestGmailSender_SendHTMLSuccess(t *testing.T) {
	var capturedHost string
	var capturedPort int
	var capturedTLSConfig *tls.Config
	var sentMessages int

	withFakeDialer(t, func(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer {
		capturedHost = host
		capturedPort = port
		capturedTLSConfig = tlsConfig
		return &fakeSMTPDialer{
			sendFunc: func(messages ...*gomail.Message) error {
				sentMessages = len(messages)
				return nil
			},
		}
	})

	sender := NewGmailSender(config.SMTPConfig{
		Host:     "smtp.gmail.com",
		Port:     587,
		Username: "user@gmail.com",
		Password: "app-password",
		From:     "VCS-SMS <user@gmail.com>",
	})

	err := sender.SendHTML(context.Background(), "admin@test.com", "Subject", "<h1>Body</h1>")

	assert.NoError(t, err)
	assert.Equal(t, "smtp.gmail.com", capturedHost)
	assert.Equal(t, 587, capturedPort)
	assert.NotNil(t, capturedTLSConfig)
	assert.Equal(t, uint16(tls.VersionTLS12), capturedTLSConfig.MinVersion)
	assert.Equal(t, 1, sentMessages)
}

func TestGmailSender_SendHTMLError(t *testing.T) {
	withFakeDialer(t, func(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer {
		return &fakeSMTPDialer{
			sendFunc: func(messages ...*gomail.Message) error {
				return fmt.Errorf("auth failed")
			},
		}
	})

	sender := NewGmailSender(config.SMTPConfig{Host: "smtp.gmail.com", Port: 587})

	err := sender.SendHTML(context.Background(), "admin@test.com", "Subject", "<h1>Body</h1>")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auth failed")
}

func TestGmailSender_SendHTMLContextCanceled(t *testing.T) {
	dialerCreated := false
	withFakeDialer(t, func(host string, port int, username, password string, tlsConfig *tls.Config) smtpDialer {
		dialerCreated = true
		return &fakeSMTPDialer{}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sender := NewGmailSender(config.SMTPConfig{Host: "smtp.gmail.com", Port: 587})

	err := sender.SendHTML(ctx, "admin@test.com", "Subject", "<h1>Body</h1>")

	assert.ErrorIs(t, err, context.Canceled)
	assert.False(t, dialerCreated)
}

func TestCreateSMTPDialerDefault(t *testing.T) {
	original := newSMTPDialerOverride
	newSMTPDialerOverride = nil
	t.Cleanup(func() {
		newSMTPDialerOverride = original
	})

	dialer := createSMTPDialer("smtp.gmail.com", 587, "user", "pass", &tls.Config{ServerName: "smtp.gmail.com"})

	assert.NotNil(t, dialer)
}
