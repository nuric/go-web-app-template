package email

import "github.com/rs/zerolog/log"

type Emailer interface {
	SendEmail(to string, subject string, body string) error
}

type LogEmailer struct{}

func (c LogEmailer) SendEmail(to string, subject string, body string) error {
	// Simulate sending email by logging it
	log.Info().Str("to", to).Str("subject", subject).Str("body", body).Msg("Sending email")
	return nil
}
