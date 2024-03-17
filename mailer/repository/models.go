package repository

import "github.com/google/uuid"

type VerEntity struct {
	Email    string
	Password string
	VerId    uuid.UUID
}
