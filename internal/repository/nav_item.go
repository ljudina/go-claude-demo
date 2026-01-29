package repository

import (
	"context"
	"database/sql"
	"errors"

	"claude-test/internal/domain"
)

type NavItemRepository struct {
	db      *sql.DB
	queries *Queries
}

func NewNavItemRepository(db *sql.DB) *NavItemRepository {
	return &NavItemRepository{
		db:      db,
		queries: New(db),
	}
}

func (r *NavItemRepository) Create(ctx context.Context, item *domain.NavItem) (*domain.NavItem, error) {
	dbItem, err := r.queries.CreateNavItem(ctx, CreateNavItemParams{
		Name:      item.Name,
		Url:       item.URL,
		Icon:      sql.NullString{String: item.Icon, Valid: item.Icon != ""},
		ParentID:  toNullInt64(item.ParentID),
		SortOrder: sql.NullInt64{Int64: int64(item.SortOrder), Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return toDomainNavItem(dbItem), nil
}

func (r *NavItemRepository) GetByID(ctx context.Context, id int64) (*domain.NavItem, error) {
	dbItem, err := r.queries.GetNavItemByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNavItemNotFound
		}
		return nil, err
	}
	return toDomainNavItem(dbItem), nil
}

func (r *NavItemRepository) GetByName(ctx context.Context, name string) (*domain.NavItem, error) {
	dbItem, err := r.queries.GetNavItemByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNavItemNotFound
		}
		return nil, err
	}
	return toDomainNavItem(dbItem), nil
}

func (r *NavItemRepository) List(ctx context.Context) ([]*domain.NavItem, error) {
	dbItems, err := r.queries.ListNavItems(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*domain.NavItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = toDomainNavItem(dbItem)
	}
	return items, nil
}

func (r *NavItemRepository) ListTopLevel(ctx context.Context) ([]*domain.NavItem, error) {
	dbItems, err := r.queries.ListTopLevelNavItems(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*domain.NavItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = toDomainNavItem(dbItem)
	}
	return items, nil
}

func (r *NavItemRepository) ListChildren(ctx context.Context, parentID int64) ([]*domain.NavItem, error) {
	dbItems, err := r.queries.ListNavItemChildren(ctx, sql.NullInt64{Int64: parentID, Valid: true})
	if err != nil {
		return nil, err
	}
	items := make([]*domain.NavItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = toDomainNavItem(dbItem)
	}
	return items, nil
}

func (r *NavItemRepository) Update(ctx context.Context, item *domain.NavItem) (*domain.NavItem, error) {
	dbItem, err := r.queries.UpdateNavItem(ctx, UpdateNavItemParams{
		ID:        item.ID,
		Name:      item.Name,
		Url:       item.URL,
		Icon:      sql.NullString{String: item.Icon, Valid: item.Icon != ""},
		ParentID:  toNullInt64(item.ParentID),
		SortOrder: sql.NullInt64{Int64: int64(item.SortOrder), Valid: true},
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNavItemNotFound
		}
		return nil, err
	}
	return toDomainNavItem(dbItem), nil
}

func (r *NavItemRepository) Delete(ctx context.Context, id int64) error {
	return r.queries.DeleteNavItem(ctx, id)
}

func (r *NavItemRepository) AssignRole(ctx context.Context, navItemID, roleID int64) error {
	return r.queries.AssignRoleToNavItem(ctx, AssignRoleToNavItemParams{
		NavItemID: navItemID,
		RoleID:    roleID,
	})
}

func (r *NavItemRepository) RemoveRole(ctx context.Context, navItemID, roleID int64) error {
	return r.queries.RemoveRoleFromNavItem(ctx, RemoveRoleFromNavItemParams{
		NavItemID: navItemID,
		RoleID:    roleID,
	})
}

func (r *NavItemRepository) ClearRoles(ctx context.Context, navItemID int64) error {
	return r.queries.ClearNavItemRoles(ctx, navItemID)
}

func (r *NavItemRepository) GetRoles(ctx context.Context, navItemID int64) ([]*domain.Role, error) {
	dbRoles, err := r.queries.GetNavItemRoles(ctx, navItemID)
	if err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}

func (r *NavItemRepository) SetRoles(ctx context.Context, navItemID int64, roleIDs []int64) error {
	if err := r.queries.ClearNavItemRoles(ctx, navItemID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err := r.queries.AssignRoleToNavItem(ctx, AssignRoleToNavItemParams{
			NavItemID: navItemID,
			RoleID:    roleID,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *NavItemRepository) GetVisibleForRoles(ctx context.Context, roleIDs []int64) ([]*domain.NavItem, error) {
	if len(roleIDs) == 0 {
		return []*domain.NavItem{}, nil
	}
	dbItems, err := r.queries.GetNavItemsForRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}
	items := make([]*domain.NavItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = toDomainNavItem(dbItem)
	}
	return items, nil
}

func toDomainNavItem(n NavItem) *domain.NavItem {
	item := &domain.NavItem{
		ID:        n.ID,
		Name:      n.Name,
		URL:       n.Url,
		SortOrder: int(n.SortOrder.Int64),
	}
	if n.Icon.Valid {
		item.Icon = n.Icon.String
	}
	if n.ParentID.Valid {
		parentID := n.ParentID.Int64
		item.ParentID = &parentID
	}
	if n.CreatedAt.Valid {
		item.CreatedAt = n.CreatedAt.Time
	}
	if n.UpdatedAt.Valid {
		item.UpdatedAt = n.UpdatedAt.Time
	}
	return item
}

func toNullInt64(ptr *int64) sql.NullInt64 {
	if ptr == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *ptr, Valid: true}
}
