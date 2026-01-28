package authz

import (
	"context"

	"claude-test/internal/domain"
	"claude-test/internal/service"
)

// RoleService wraps the role service to implement RoleGetter
type RoleService struct {
	svc *service.RoleService
}

// NewRoleService creates a new RoleService wrapper
func NewRoleService(svc *service.RoleService) *RoleService {
	return &RoleService{svc: svc}
}

// GetUserRoles returns roles for a user
func (s *RoleService) GetUserRoles(userID int64) ([]*domain.Role, error) {
	return s.svc.GetUserRoles(context.Background(), userID)
}
