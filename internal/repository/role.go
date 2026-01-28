package repository

import (
	"context"
	"database/sql"
	"errors"

	"claude-test/internal/domain"
)

type RoleRepository struct {
	db      *sql.DB
	queries *Queries
}

func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{
		db:      db,
		queries: New(db),
	}
}

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	dbRole, err := r.queries.CreateRole(ctx, CreateRoleParams{
		Name:        role.Name,
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
	})
	if err != nil {
		return nil, err
	}
	return toDomainRole(dbRole), nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id int64) (*domain.Role, error) {
	dbRole, err := r.queries.GetRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return toDomainRole(dbRole), nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	dbRole, err := r.queries.GetRoleByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return toDomainRole(dbRole), nil
}

func (r *RoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	dbRoles, err := r.queries.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	dbRole, err := r.queries.UpdateRole(ctx, UpdateRoleParams{
		ID:          role.ID,
		Name:        role.Name,
		Description: sql.NullString{String: role.Description, Valid: role.Description != ""},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return toDomainRole(dbRole), nil
}

func (r *RoleRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteRole(ctx, id)
}

func (r *RoleRepository) AssignToUser(ctx context.Context, userID, roleID int64) error {
	return r.queries.AssignRoleToUser(ctx, AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *RoleRepository) RemoveFromUser(ctx context.Context, userID, roleID int64) error {
	return r.queries.RemoveRoleFromUser(ctx, RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	})
}

func (r *RoleRepository) GetUserRoles(ctx context.Context, userID int64) ([]*domain.Role, error) {
	dbRoles, err := r.queries.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *RoleRepository) SetUserRoles(ctx context.Context, userID int64, roleIDs []int64) error {
	if err := r.queries.ClearUserRoles(ctx, userID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err := r.queries.AssignRoleToUser(ctx, AssignRoleToUserParams{
			UserID: userID,
			RoleID: roleID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func toDomainRole(r Role) *domain.Role {
	return &domain.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description.String,
	}
}
