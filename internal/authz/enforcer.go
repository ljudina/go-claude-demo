package authz

import (
	"database/sql"
	"embed"
	"fmt"

	sqladapter "github.com/Blank-Xu/sql-adapter"
	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
)

//go:embed model.conf
var modelFS embed.FS

// NewEnforcer creates a new Casbin enforcer with SQLite adapter
func NewEnforcer(db *sql.DB) (*casbin.Enforcer, error) {
	// Load model from embedded file
	modelData, err := modelFS.ReadFile("model.conf")
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	m, err := model.NewModelFromString(string(modelData))
	if err != nil {
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Create SQL adapter
	adapter, err := sqladapter.NewAdapter(db, "sqlite3", "casbin_rules")
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create enforcer: %w", err)
	}

	// Load policies from database
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("failed to load policy: %w", err)
	}

	return enforcer, nil
}

// SetupDefaultPolicies creates default policies if none exist
func SetupDefaultPolicies(e *casbin.Enforcer) error {
	policies, _ := e.GetPolicy()
	if len(policies) > 0 {
		return nil // Policies already exist
	}

	// Default policies for Admin role
	adminPolicies := [][]string{
		{"Admin", "/web/users", "read"},
		{"Admin", "/web/users", "write"},
		{"Admin", "/web/users/*", "read"},
		{"Admin", "/web/users/*", "write"},
		{"Admin", "/web/roles", "read"},
		{"Admin", "/web/roles", "write"},
		{"Admin", "/web/roles/*", "read"},
		{"Admin", "/web/roles/*", "write"},
		{"Admin", "/api/users", "read"},
		{"Admin", "/api/users", "write"},
		{"Admin", "/api/users/*", "read"},
		{"Admin", "/api/users/*", "write"},
		{"Admin", "/api/roles", "read"},
		{"Admin", "/api/roles", "write"},
		{"Admin", "/api/roles/*", "read"},
		{"Admin", "/api/roles/*", "write"},
	}

	// Default policies for User role (read-only access to own profile)
	userPolicies := [][]string{
		{"User", "/", "read"},
	}

	for _, p := range adminPolicies {
		if _, err := e.AddPolicy(p); err != nil {
			return fmt.Errorf("failed to add admin policy: %w", err)
		}
	}

	for _, p := range userPolicies {
		if _, err := e.AddPolicy(p); err != nil {
			return fmt.Errorf("failed to add user policy: %w", err)
		}
	}

	return e.SavePolicy()
}
