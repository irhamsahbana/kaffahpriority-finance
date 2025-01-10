package cmd

import (
	"codebase-app/internal/infrastructure/config"
	"encoding/json"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog/log"
	"gopkg.in/gomail.v2"
)

type ForgotPasswordEventPayload struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	Token string `json:"token"`
}

func ForgotPasswordHandler(msg jetstream.Msg) {
	payload := &ForgotPasswordEventPayload{}

	err := json.Unmarshal(msg.Data(), payload)
	if err != nil {
		log.Error().Err(err).Msg("consumer::ForgotPasswordHandler Error while unmarshalling payload")
		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Msg("consumer::ForgotPasswordHandler Error while rejecting message")
		}
		return
	}

	err = sendForgotPassword([]string{payload.Email}, payload)
	if err != nil {
		log.Error().Err(err).Any("payload", payload).Msg("consumer::ForgotPasswordHandler Error while sending email")
		if err := msg.Ack(); err != nil {
			log.Error().Err(err).Any("payload", payload).Msg("consumer::ForgotPasswordHandler Error while rejecting message")
		}
		return
	}

	if err := msg.Ack(); err != nil {
		log.Error().Err(err).Any("payload", payload).Msg("consumer::ForgotPasswordHandler Error while acknowledging message")
	}
}

func sendForgotPassword(to []string, payload *ForgotPasswordEventPayload) error {
	var (
		mail             = config.Envs.Mail
		feUrl            = config.Envs.FrontendURL
		mailSmtpHost     = mail.Host
		mailSmtpPort     = mail.Port
		mailSmtpUsername = mail.Username
		mailSmtpPassword = mail.Password
		baseUrl          string
	)

	if payload.Role == "client" {
		baseUrl = feUrl.ClientBaseURL
	} else {
		baseUrl = feUrl.AdminBaseURL
	}
	body := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Password Reset</title>
		</head>
		<body>
			<p>Click the link below to reset your password</p>
			<a href="` + baseUrl + feUrl.PasswordReset + `?` + payload.Token + `">Reset Password</a>
		</body>
		</html>
	`

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", "Crowners <"+mailSmtpUsername+">")
	mailer.SetHeader("To", to...)
	mailer.SetHeader("Subject", "Password Reset")
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(mailSmtpHost, mailSmtpPort, mailSmtpUsername, mailSmtpPassword)
	err := dialer.DialAndSend(mailer)
	if err != nil {
		log.Error().Err(err).Msg("consumer::sendForgotPassword Error while sending email")
		return err
	}

	return nil
}
