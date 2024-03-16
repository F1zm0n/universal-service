package models

import "github.com/google/uuid"

type VerDto struct {
	VerID    uuid.UUID `json:"ver_id,omitempty"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}
