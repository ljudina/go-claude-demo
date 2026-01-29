package domain

import (
	"context"
	"strings"
	"time"
)

type NavItem struct {
	ID        int64
	Name      string
	URL       string
	Icon      string
	ParentID  *int64
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
	Children  []*NavItem
	Roles     []*Role
}

func (n *NavItem) Validate() error {
	if strings.TrimSpace(n.Name) == "" {
		return ErrInvalidNavItemName
	}
	if strings.TrimSpace(n.URL) == "" {
		return ErrInvalidNavItemURL
	}
	return nil
}

func (n *NavItem) HasChildren() bool {
	return len(n.Children) > 0
}

type NavItemRepository interface {
	Create(ctx context.Context, item *NavItem) (*NavItem, error)
	GetByID(ctx context.Context, id int64) (*NavItem, error)
	GetByName(ctx context.Context, name string) (*NavItem, error)
	List(ctx context.Context) ([]*NavItem, error)
	ListTopLevel(ctx context.Context) ([]*NavItem, error)
	ListChildren(ctx context.Context, parentID int64) ([]*NavItem, error)
	Update(ctx context.Context, item *NavItem) (*NavItem, error)
	Delete(ctx context.Context, id int64) error
	AssignRole(ctx context.Context, navItemID, roleID int64) error
	RemoveRole(ctx context.Context, navItemID, roleID int64) error
	ClearRoles(ctx context.Context, navItemID int64) error
	GetRoles(ctx context.Context, navItemID int64) ([]*Role, error)
	SetRoles(ctx context.Context, navItemID int64, roleIDs []int64) error
	GetVisibleForRoles(ctx context.Context, roleIDs []int64) ([]*NavItem, error)
}
