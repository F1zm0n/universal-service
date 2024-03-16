package service

import (
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

const (
	smtpAuthAddres   = "smtp.gmail.com"
	smtpServerAddres = "smtp.gmail.com:587"
)

type MailSender interface {
	SendMail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddres   string
	fromEmailPassword string
}

func NewGmailSender() MailSender {
	return &GmailSender{
		name:              viper.GetString("gmail.name"),
		fromEmailAddres:   viper.GetString("gmail.add"),
		fromEmailPassword: viper.GetString("gmail.pass"),
	}
}

func (s GmailSender) SendMail(
	subject string,
	content string,
	to []string,
	cc []string,
	bcc []string,
	attachFiles []string,
) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.name, s.fromEmailAddres)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc
	for _, f := range attachFiles {
		_, err := e.AttachFile(f)
		if err != nil {
			return err
		}
	}

	smtpAuth := smtp.PlainAuth("", s.fromEmailAddres, s.fromEmailPassword, smtpAuthAddres)
	return e.Send(smtpServerAddres, smtpAuth)
}
