package service

import (
	"context"

	"claude-test/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, email, name string) (*domain.User, error) {
	user := &domain.User{
		Email: email,
		Name:  name,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	return s.repo.Create(ctx, user)
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) Update(ctx context.Context, id int64, email, name string) (*domain.User, error) {
	user := &domain.User{
		ID:    id,
		Email: email,
		Name:  name,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByEmail(ctx, email)
	if err == nil && existing != nil && existing.ID != id {
		return nil, domain.ErrUserAlreadyExists
	}

	return s.repo.Update(ctx, user)
}

func (s *UserService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *UserService) UpdateTheme(ctx context.Context, id int64, theme string) (*domain.User, error) {
	if theme != "light" && theme != "dark" {
		theme = "light"
	}
	return s.repo.UpdateTheme(ctx, id, theme)
}
