package email

import (
	"context"

	"gitlab.misakey.dev/misakey/backend/api/src/sdk/logger"
)

// MailerLogger logs emails instead of sending them.
type MailerLogger struct {
}

// NewLogMailer creates a mailer that doesn't deliver emails but simply logs them.
func NewLogMailer() *MailerLogger {
	return &MailerLogger{}
}

// Send an email (log only text)
func (l MailerLogger) Send(ctx context.Context, email *Notification) error {
	logger.
		FromCtx(ctx).
		Info().
		Msgf("===> EMAIL SENT TO %s FROM %s: [%s]", email.To, email.From, email.Subject)
	logger.
		FromCtx(ctx).
		Info().
		Msg(email.TextBody)
	return nil
}
