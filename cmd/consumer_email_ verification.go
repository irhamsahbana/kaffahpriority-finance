package cmd

import (
	"codebase-app/internal/infrastructure/config"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type EmailVerificationEventPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

func EmailVerificationHandler(msg jetstream.Msg) {
	payload := &EmailVerificationEventPayload{}

	err := json.Unmarshal(msg.Data(), payload)
	if err != nil {
		log.Error().Err(err).Msg("consumer::EmailVerificationHandler Error while unmarshalling payload")
		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Msg("consumer::EmailVerificationHandler Error while rejecting message")
		}
		return
	}

	log.Debug().Any("payload", payload).Msg("consumer::EmailVerificationHandler Received message")

	err = sendEmailVerification([]string{payload.Email}, payload)
	if err != nil {
		log.Error().Err(err).Any("payload", payload).Msg("consumer::EmailVerificationHandler Error while sending email")
		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Any("payload", payload).Msg("consumer::EmailVerificationHandler Error while rejecting message")
		}
		return
	}
	if err := msg.Ack(); err != nil {
		log.Error().Err(err).Any("payload", payload).Msg("consumer::EmailVerificationHandler Error while acknowledging message")
	}
}

func sendEmailVerification(to []string, payload *EmailVerificationEventPayload) error {

	var (
		mail             = config.Envs.Mail
		feUrl            = config.Envs.FrontendURL
		mailSmtpHost     = mail.Host
		mailSmtpPort     = mail.Port
		mailSmtpUsername = mail.Username
		mailSmtpPassword = mail.Password
	)

	body := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Email Verification</title>
		</head>
		<body>
			<p>Click the link below to verify your email</p>
			<a href="` + feUrl.ClientBaseURL + feUrl.EmailVerification + `?` + payload.Token + `">Verify Email</a>
		</body>
		</html>
	`

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "Crowners <"+mailSmtpUsername+">")
	mailer.SetHeader("To", to...)
	mailer.SetHeader("Subject", "Email Verification")
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(mailSmtpHost, mailSmtpPort, mailSmtpUsername, mailSmtpPassword)
	err := dialer.DialAndSend(mailer)
	if err != nil {
		log.Error().Err(err).Msg("consumer::sendEmailVerification Error while sending email")
		return err
	}

	return nil
}
