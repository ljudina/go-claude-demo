package authz

import (
	"net/http"

	"github.com/casbin/casbin/v3"

	"claude-test/internal/domain"
)

// UserGetter is an interface to get current user from request
type UserGetter interface {
	GetCurrentUser(r *http.Request) *domain.User
}

// RoleGetter is an interface to get user roles
type RoleGetter interface {
	GetUserRoles(userID int64) ([]*domain.Role, error)
}

// Middleware creates authorization middleware
type Middleware struct {
	enforcer   *casbin.Enforcer
	userGetter UserGetter
	roleGetter RoleGetter
}

// NewMiddleware creates a new authorization middleware
func NewMiddleware(enforcer *casbin.Enforcer, userGetter UserGetter, roleGetter RoleGetter) *Middleware {
	return &Middleware{
		enforcer:   enforcer,
		userGetter: userGetter,
		roleGetter: roleGetter,
	}
}

// RequireAuth ensures user is authenticated
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := m.userGetter.GetCurrentUser(r)
		if user == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePermission checks if user has permission to access resource
func (m *Middleware) RequirePermission(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := m.userGetter.GetCurrentUser(r)
			if user == nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			// Get user roles
			roles, err := m.roleGetter.GetUserRoles(user.ID)
			if err != nil {
				http.Error(w, "Failed to get user roles", http.StatusInternalServerError)
				return
			}

			// Check permission for each role
			path := r.URL.Path
			allowed := false
			for _, role := range roles {
				ok, err := m.enforcer.Enforce(role.Name, path, action)
				if err != nil {
					http.Error(w, "Authorization error", http.StatusInternalServerError)
					return
				}
				if ok {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if user has specific role
func (m *Middleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := m.userGetter.GetCurrentUser(r)
			if user == nil {
				http.Redirect(w, r, "/", http.StatusFound)
				return
			}

			roles, err := m.roleGetter.GetUserRoles(user.ID)
			if err != nil {
				http.Error(w, "Failed to get user roles", http.StatusInternalServerError)
				return
			}

			hasRole := false
			for _, role := range roles {
				if role.Name == roleName {
					hasRole = true
					break
				}
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
