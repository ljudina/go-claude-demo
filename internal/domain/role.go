package domain

import (
	"context"
	"strings"
)

type Role struct {
	ID          int64
	Name        string
	Description string
}

func (r *Role) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrInvalidRoleName
	}
	return nil
}

type RoleRepository interface {
	Create(ctx context.Context, role *Role) (*Role, error)
	GetByID(ctx context.Context, id int64) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	Update(ctx context.Context, role *Role) (*Role, error)
	Delete(ctx context.Context, id int64) error
	AssignToUser(ctx context.Context, userID, roleID int64) error
	RemoveFromUser(ctx context.Context, userID, roleID int64) error
	GetUserRoles(ctx context.Context, userID int64) ([]*Role, error)
	SetUserRoles(ctx context.Context, userID int64, roleIDs []int64) error
}
