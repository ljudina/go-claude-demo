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
	userService *service.UserService
	roleRepo    domain.RoleRepository
	store       *sessions.CookieStore
}

func NewAuthHandler(userService *service.UserService, roleRepo domain.RoleRepository, sessionSecret string) *AuthHandler {
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
		auth = &templates.AuthInfo{
			User:    user,
			Roles:   roles,
			IsAdmin: isAdmin,
		}
	}

	templates.HomePage(user, auth).Render(r.Context(), w)
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
