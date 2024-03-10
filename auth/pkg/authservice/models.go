package authservice

import (
	"net/mail"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
}

func validateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
