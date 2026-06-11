package mocks

import "context"

// EmailSenderMock is a mock implementation of email.EmailSender.
type EmailSenderMock struct {
	SendHTMLFunc func(ctx context.Context, to, subject, htmlBody string) error
}

func (m *EmailSenderMock) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	if m.SendHTMLFunc != nil {
		return m.SendHTMLFunc(ctx, to, subject, htmlBody)
	}
	return nil
}
