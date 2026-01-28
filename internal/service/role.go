package service

import (
	"context"

	"claude-test/internal/domain"
)

type RoleService struct {
	repo domain.RoleRepository
}

func NewRoleService(repo domain.RoleRepository) *RoleService {
	return &RoleService{repo: repo}
}

func (s *RoleService) Create(ctx context.Context, name, description string) (*domain.Role, error) {
	role := &domain.Role{
		Name:        name,
		Description: description,
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, domain.ErrRoleAlreadyExists
	}

	return s.repo.Create(ctx, role)
}

func (s *RoleService) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RoleService) List(ctx context.Context) ([]*domain.Role, error) {
	return s.repo.List(ctx)
}

func (s *RoleService) Update(ctx context.Context, id int64, name, description string) (*domain.Role, error) {
	role := &domain.Role{
		ID:          id,
		Name:        name,
		Description: description,
	}

	if err := role.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByName(ctx, name)
	if err == nil && existing != nil && existing.ID != id {
		return nil, domain.ErrRoleAlreadyExists
	}

	return s.repo.Update(ctx, role)
}

func (s *RoleService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *RoleService) GetUserRoles(ctx context.Context, userID int64) ([]*domain.Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *RoleService) SetUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	return s.repo.SetUserRoles(ctx, userID, roleIDs)
}
