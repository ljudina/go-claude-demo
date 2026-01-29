package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"

	"claude-test/internal/domain"
	"claude-test/internal/handler/templates"
	"claude-test/internal/service"
)

type AuthHandler struct {
	userService    *service.UserService
	roleRepo       domain.RoleRepository
	navItemRepo    domain.NavItemRepository
	store          *sessions.CookieStore
}

func NewAuthHandler(userService *service.UserService, roleRepo domain.RoleRepository, navItemRepo domain.NavItemRepository, sessionSecret string) *AuthHandler {
	store := sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
	}
	gothic.Store = store

	return &AuthHandler{
		userService: userService,
		roleRepo:    roleRepo,
		navItemRepo: navItemRepo,
		store:       store,
	}
}

func (h *AuthHandler) InitGoth(clientID, clientSecret, callbackURL string) {
	goth.UseProviders(
		google.New(clientID, clientSecret, callbackURL+"/auth/google/callback", "email", "profile"),
	)
}

func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.Home)
	r.Get("/auth/{provider}", h.BeginAuth)
	r.Get("/auth/{provider}/callback", h.Callback)
	r.Get("/logout", h.Logout)
	r.Put("/api/theme", h.ToggleTheme)
}

func (h *AuthHandler) Home(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "auth-session")
	user, _ := session.Values["user"].(*domain.User)

	var auth *templates.AuthInfo
	if user != nil {
		roles, _ := h.roleRepo.GetUserRoles(r.Context(), user.ID)
		isAdmin := false
		for _, role := range roles {
			if role.Name == "Admin" {
				isAdmin = true
				break
			}
		}

		// Get nav items for user's roles
		roleIDs := make([]int64, len(roles))
		for i, role := range roles {
			roleIDs[i] = role.ID
		}
		navItems, _ := h.navItemRepo.GetVisibleForRoles(r.Context(), roleIDs)
		navTree := buildNavTree(navItems)

		auth = &templates.AuthInfo{
			User:     user,
			Roles:    roles,
			IsAdmin:  isAdmin,
			NavItems: navTree,
		}
	}

	templates.HomePage(user, auth).Render(r.Context(), w)
}

// buildNavTree converts a flat list of nav items into a tree structure
func buildNavTree(items []*domain.NavItem) []*domain.NavItem {
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

func (h *AuthHandler) BeginAuth(w http.ResponseWriter, r *http.Request) {
	gothic.BeginAuthHandler(w, r)
}

func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	// Check if user exists
	user, err := h.userService.GetByEmail(ctx, gothUser.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			// Create new user
			user, err = h.userService.Create(ctx, gothUser.Email, gothUser.Name)
			if err != nil {
				http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Assign default "User" role
			if err := h.assignDefaultRole(ctx, user.ID); err != nil {
				// Log error but don't fail login
			}
		} else {
			http.Error(w, "Failed to get user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Save user to session
	session, _ := h.store.Get(r, "auth-session")
	session.Values["user"] = user
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "auth-session")
	session.Values["user"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)

	gothic.Logout(w, r)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) assignDefaultRole(ctx context.Context, userID int64) error {
	role, err := h.roleRepo.GetByName(ctx, "User")
	if err != nil {
		return err
	}
	return h.roleRepo.AssignToUser(ctx, userID, role.ID)
}

// GetCurrentUser returns the current user from session (for use in middleware)
func (h *AuthHandler) GetCurrentUser(r *http.Request) *domain.User {
	session, _ := h.store.Get(r, "auth-session")
	user, ok := session.Values["user"].(*domain.User)
	if !ok {
		return nil
	}
	return user
}

// ToggleTheme toggles the user's theme between light and dark
func (h *AuthHandler) ToggleTheme(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, "auth-session")
	user, ok := session.Values["user"].(*domain.User)
	if !ok || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Toggle theme
	newTheme := "dark"
	if user.Theme == "dark" {
		newTheme = "light"
	}

	// Update in database
	updatedUser, err := h.userService.UpdateTheme(r.Context(), user.ID, newTheme)
	if err != nil {
		http.Error(w, "Failed to update theme", http.StatusInternalServerError)
		return
	}

	// Update session
	session.Values["user"] = updatedUser
	session.Save(r, w)

	// Return the new theme
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"theme":"` + newTheme + `"}`))
}
