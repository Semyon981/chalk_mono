package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer interface {
	SendMail(From string, To string, message []byte) error
}

func New(smtpHost string, smtpPort string, smtpAuthUsername string, smtpAuthPassword string) Mailer {
	return &mailer{
		smtpHost:         smtpHost,
		smtpPort:         smtpPort,
		smtpAuthUsername: smtpAuthUsername,
		smtpAuthPassword: smtpAuthPassword,
	}
}

type mailer struct {
	smtpHost         string
	smtpPort         string
	smtpAuthUsername string
	smtpAuthPassword string
}

func (m *mailer) SendMail(From string, To string, message []byte) error {
	auth := smtp.PlainAuth("", m.smtpAuthUsername, m.smtpAuthPassword, m.smtpHost)
	err := smtp.SendMail(m.smtpHost+":"+m.smtpPort, auth, From, []string{To}, message)
	if err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}
	return nil
}
