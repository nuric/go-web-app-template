package email

import (
	"log/slog"
)

type Emailer interface {
	SendEmail(to string, subject string, body string) error
}

type LogEmailer struct{}

func (c LogEmailer) SendEmail(to string, subject string, body string) error {
	// Simulate sending email by logging it
	slog.Debug("Sending email", "to", to, "subject", subject, "body", body)
	return nil
}
