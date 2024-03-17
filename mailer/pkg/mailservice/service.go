package mailservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/go-kit/log"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/spf13/viper"

	"github.com/F1zm0n/universal-mailer/repository"
)

func New(log log.Logger) Service {
	var svc Service
	{
		svc = NewBaseService()
		svc = LoggingMiddleware(log)(svc)
		svc = InstrumentingMiddleware()(svc)
	}
	return svc
}

type Service interface {
	SendEmail(ctx context.Context, ver VerDto) error
	VerifyMail(ctx context.Context, id uuid.UUID) error
}

type baseService struct {
	mailer MailSender
	db     repository.Repository
}

type VerDto struct {
	VerID    uuid.UUID `json:"ver_id,omitempty"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}
type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewBaseService() Service {
	db := repository.NewPostgresRepository()
	mailer := NewGmailSender()
	return &baseService{
		db:     db,
		mailer: mailer,
	}
}

func (s baseService) VerifyMail(ctx context.Context, id uuid.UUID) error {
	user, err := s.db.GetByVerId(id)
	if err != nil {
		return err
	}
	// authLink := fmt.Sprintf("%s%s",
	// 	viper.GetString("req.auth.addres"),
	// 	viper.GetString("req.auth.create"),
	// )
	payload := AuthPayload{
		Email:    user.Email,
		Password: user.Password,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		"http://auth:5000/register",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	tx, err := s.db.DeleteByVerId(id)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer tx.Commit()
	cli := http.DefaultClient
	res, err := cli.Do(req)
	if err != nil {
		tx.Rollback()
		return err
	}
	if res.StatusCode != 200 {
		tx.Rollback()
		return fmt.Errorf("error status code is not 200:%s", res.Status)
	}
	return nil
}

func (s baseService) SendEmail(ctx context.Context, ver VerDto) error {
	user := repository.VerEntity{
		VerId:    uuid.New(),
		Email:    ver.Email,
		Password: ver.Password,
	}
	tx, err := s.db.CreateLink(user)
	if err != nil {
		tx.Rollback()
		return err
	}

	subject := "verify your email addres"
	// content := fmt.Sprintf(`
	// 	<h1>verify your email</h1>
	// 	<p>Link<a href="http://%s%s?id=%v">Verify</a></p>
	// `, viper.GetString("req.http.mailer.addres"),
	// 	viper.GetString("req.http.mailer.verify"),
	// 	user.VerId,
	// )
	content := fmt.Sprintf(`
		<h1>verify your email</h1>
		<p>Link<a href="http://%s%s?id=%v">Verify</a></p>
		`, "localhost:5002",
		"/verify",
		user.VerId,
	)
	to := []string{user.Email}
	err = s.mailer.SendMail(subject, content, to, nil, nil, nil)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

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
