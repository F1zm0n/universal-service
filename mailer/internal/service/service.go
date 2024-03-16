package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/spf13/viper"

	models "github.com/F1zm0n/universal-mailer/internal"
	"github.com/F1zm0n/universal-mailer/internal/repository"
)

type Service interface {
	SendEmail(ver models.VerDto) error
	VerifyMail(id uuid.UUID) error
}

type baseService struct {
	mailer MailSender
	db     repository.Repository
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

func (s baseService) VerifyMail(id uuid.UUID) error {
	user, err := s.db.GetByVerId(id)
	if err != nil {
		log.Println("err ", err)
		return err
	}
	err = s.db.DeleteByVerId(id)
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
	req, err := http.NewRequest(http.MethodPost, "http://auth:8081/register", bytes.NewReader(b))
	if err != nil {
		return err
	}
	cli := http.DefaultClient
	res, err := cli.Do(req)
	if err != nil {
		return err
	}
	log.Println("sent request")
	if res.StatusCode != 200 {
		return fmt.Errorf("error status code is not 200:%s", res.Status)
	}

	return nil
}

func (s baseService) SendEmail(ver models.VerDto) error {
	user := repository.VerEntity{
		VerId:    uuid.New(),
		Email:    ver.Email,
		Password: ver.Password,
	}

	subject := "verify your email addres"
	content := fmt.Sprintf(`
		<h1>verify your email</h1>
		<p>Link<a href="http://%s%s?id=%v">Verify</a></p>
	`, viper.GetString("req.http.mailer.addres"),
		viper.GetString("req.http.mailer.verify"),
		user.VerId,
	)
	to := []string{user.Email}
	err := s.mailer.SendMail(subject, content, to, nil, nil, nil)
	if err != nil {
		return err
	}
	err = s.db.CreateLink(user)
	if err != nil {
		return err
	}
	return nil
}
