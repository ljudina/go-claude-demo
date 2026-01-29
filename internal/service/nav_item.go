package service

import (
	"context"

	"claude-test/internal/domain"
)

type NavItemService struct {
	repo        domain.NavItemRepository
	roleService *RoleService
}

func NewNavItemService(repo domain.NavItemRepository, roleService *RoleService) *NavItemService {
	return &NavItemService{
		repo:        repo,
		roleService: roleService,
	}
}

func (s *NavItemService) Create(ctx context.Context, name, url, icon string, parentID *int64, sortOrder int, roleIDs []int64) (*domain.NavItem, error) {
	item := &domain.NavItem{
		Name:      name,
		URL:       url,
		Icon:      icon,
		ParentID:  parentID,
		SortOrder: sortOrder,
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	created, err := s.repo.Create(ctx, item)
	if err != nil {
		return nil, err
	}

	if len(roleIDs) > 0 {
		if err := s.repo.SetRoles(ctx, created.ID, roleIDs); err != nil {
			return nil, err
		}
	}

	return created, nil
}

func (s *NavItemService) GetByID(ctx context.Context, id int64) (*domain.NavItem, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	roles, err := s.repo.GetRoles(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Roles = roles
	return item, nil
}

func (s *NavItemService) List(ctx context.Context) ([]*domain.NavItem, error) {
	return s.repo.List(ctx)
}

func (s *NavItemService) ListWithRoles(ctx context.Context) ([]*domain.NavItem, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		roles, err := s.repo.GetRoles(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Roles = roles
	}
	return items, nil
}

func (s *NavItemService) Update(ctx context.Context, id int64, name, url, icon string, parentID *int64, sortOrder int, roleIDs []int64) (*domain.NavItem, error) {
	item := &domain.NavItem{
		ID:        id,
		Name:      name,
		URL:       url,
		Icon:      icon,
		ParentID:  parentID,
		SortOrder: sortOrder,
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	updated, err := s.repo.Update(ctx, item)
	if err != nil {
		return nil, err
	}

	if err := s.repo.SetRoles(ctx, id, roleIDs); err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *NavItemService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

func (s *NavItemService) GetRoles(ctx context.Context, navItemID int64) ([]*domain.Role, error) {
	return s.repo.GetRoles(ctx, navItemID)
}

func (s *NavItemService) SetRoles(ctx context.Context, navItemID int64, roleIDs []int64) error {
	return s.repo.SetRoles(ctx, navItemID, roleIDs)
}

// GetMenuForUser returns a tree of nav items visible to the user based on their roles
func (s *NavItemService) GetMenuForUser(ctx context.Context, userID int64) ([]*domain.NavItem, error) {
	roles, err := s.roleService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return []*domain.NavItem{}, nil
	}

	roleIDs := make([]int64, len(roles))
	for i, role := range roles {
		roleIDs[i] = role.ID
	}

	items, err := s.repo.GetVisibleForRoles(ctx, roleIDs)
	if err != nil {
		return nil, err
	}

	return s.BuildTree(items), nil
}

// BuildTree converts a flat list of nav items into a tree structure
func (s *NavItemService) BuildTree(items []*domain.NavItem) []*domain.NavItem {
	if len(items) == 0 {
		return items
	}

	// Create a map for quick lookup
	itemMap := make(map[int64]*domain.NavItem)
	for _, item := range items {
		item.Children = []*domain.NavItem{} // Initialize children slice
		itemMap[item.ID] = item
	}

	// Build the tree
	var roots []*domain.NavItem
	for _, item := range items {
		if item.ParentID == nil {
			roots = append(roots, item)
		} else {
			parent, exists := itemMap[*item.ParentID]
			if exists {
				parent.Children = append(parent.Children, item)
			} else {
				// Parent not in visible items, treat as root
				roots = append(roots, item)
			}
		}
	}

	return roots
}

// GetAllTopLevel returns all top-level nav items (for parent selection in admin)
func (s *NavItemService) GetAllTopLevel(ctx context.Context) ([]*domain.NavItem, error) {
	return s.repo.ListTopLevel(ctx)
}

// SeedDefaultNavItems creates default navigation items if none exist
func (s *NavItemService) SeedDefaultNavItems(ctx context.Context) error {
	items, err := s.repo.List(ctx)
	if err != nil {
		return err
	}
	if len(items) > 0 {
		return nil // Nav items already exist
	}

	// Get Admin role ID
	roles, err := s.roleService.List(ctx)
	if err != nil {
		return err
	}

	var adminRoleID int64
	for _, role := range roles {
		if role.Name == "Admin" {
			adminRoleID = role.ID
			break
		}
	}

	if adminRoleID == 0 {
		return nil // Admin role not found, skip seeding
	}

	adminRoleIDs := []int64{adminRoleID}

	// Create Home item (visible to Admin)
	_, err = s.Create(ctx, "Home", "/", "home", nil, 0, adminRoleIDs)
	if err != nil {
		return err
	}

	// Create Administration dropdown parent
	adminItem, err := s.Create(ctx, "Administration", "#", "settings", nil, 10, adminRoleIDs)
	if err != nil {
		return err
	}

	// Create child items under Administration
	_, err = s.Create(ctx, "Users", "/web/users", "users", &adminItem.ID, 0, adminRoleIDs)
	if err != nil {
		return err
	}

	_, err = s.Create(ctx, "Roles", "/web/roles", "shield", &adminItem.ID, 10, adminRoleIDs)
	if err != nil {
		return err
	}

	_, err = s.Create(ctx, "Policies", "/web/policies", "clipboard", &adminItem.ID, 20, adminRoleIDs)
	if err != nil {
		return err
	}

	_, err = s.Create(ctx, "Navigation", "/web/nav-items", "navigation", &adminItem.ID, 30, adminRoleIDs)
	if err != nil {
		return err
	}

	return nil
}
