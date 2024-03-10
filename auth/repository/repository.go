package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;"`
	Email     string    `gorm:"not null;unique"`
	Password  []byte    `gorm:"not null"`
	CreatedAt time.Time
}

type Repository interface {
	InsertUser(ctx context.Context, user User) error
	GetUserByEmail(ctx context.Context, email string) (User, error)
}
