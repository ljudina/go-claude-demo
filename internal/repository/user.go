package repository

import (
	"context"
	"database/sql"
	"errors"

	"claude-test/internal/domain"
)

type UserRepository struct {
	queries *Queries
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		queries: New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	theme := user.Theme
	if theme == "" {
		theme = "light"
	}
	dbUser, err := r.queries.CreateUser(ctx, CreateUserParams{
		Email: user.Email,
		Name:  user.Name,
		Theme: theme,
	})
	if err != nil {
		return nil, err
	}
	return toDomainUserFromCreate(dbUser), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	dbUser, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUserFromGetByID(dbUser), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	dbUser, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUserFromGetByEmail(dbUser), nil
}

func (r *UserRepository) List(ctx context.Context) ([]*domain.User, error) {
	dbUsers, err := r.queries.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toDomainUserFromList(dbUser)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	dbUser, err := r.queries.UpdateUser(ctx, UpdateUserParams{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUserFromUpdate(dbUser), nil
}

func (r *UserRepository) UpdateTheme(ctx context.Context, id int64, theme string) (*domain.User, error) {
	dbUser, err := r.queries.UpdateUserTheme(ctx, UpdateUserThemeParams{
		ID:    id,
		Theme: theme,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUserFromUpdateTheme(dbUser), nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteUser(ctx, id)
}

func toDomainUserFromCreate(u CreateUserRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func toDomainUserFromGetByID(u GetUserByIDRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func toDomainUserFromGetByEmail(u GetUserByEmailRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func toDomainUserFromList(u ListUsersRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func toDomainUserFromUpdate(u UpdateUserRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}

func toDomainUserFromUpdateTheme(u UpdateUserThemeRow) *domain.User {
	return &domain.User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Theme:     u.Theme,
		CreatedAt: u.CreatedAt.Time,
		UpdatedAt: u.UpdatedAt.Time,
	}
}
