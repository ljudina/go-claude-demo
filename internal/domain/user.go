package domain

import (
	"context"
	"strings"
	"time"
)

type User struct {
	ID        int64
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Email) == "" || !strings.Contains(u.Email, "@") {
		return ErrInvalidEmail
	}
	if strings.TrimSpace(u.Name) == "" {
		return ErrInvalidName
	}
	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id int64) error
}
