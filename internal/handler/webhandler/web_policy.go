package webhandler

import (
	"net/http"

	"github.com/casbin/casbin/v3"

	"claude-test/internal/domain"
	"claude-test/internal/handler"
	"claude-test/internal/handler/templates"
	"claude-test/internal/service"
)

type WebPolicyHandler struct {
	enforcer       *casbin.Enforcer
	roleService    *service.RoleService
	navItemService *service.NavItemService
	authHandler    *handler.AuthHandler
}

func NewWebPolicyHandler(enforcer *casbin.Enforcer, roleService *service.RoleService, navItemService *service.NavItemService, authHandler *handler.AuthHandler) *WebPolicyHandler {
	return &WebPolicyHandler{
		enforcer:       enforcer,
		roleService:    roleService,
		navItemService: navItemService,
		authHandler:    authHandler,
	}
}

func (h *WebPolicyHandler) getAuthInfo(r *http.Request) *templates.AuthInfo {
	user := h.authHandler.GetCurrentUser(r)
	if user == nil {
		return nil
	}

	roles, _ := h.roleService.GetUserRoles(r.Context(), user.ID)
	isAdmin := false
	for _, role := range roles {
		if role.Name == "Admin" {
			isAdmin = true
			break
		}
	}

	navItems, _ := h.navItemService.GetMenuForUser(r.Context(), user.ID)

	return &templates.AuthInfo{
		User:     user,
		Roles:    roles,
		IsAdmin:  isAdmin,
		NavItems: navItems,
	}
}

func (h *WebPolicyHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	rawPolicies, _ := h.enforcer.GetPolicy()
	policies := make([]*domain.Policy, 0, len(rawPolicies))
	for _, p := range rawPolicies {
		if len(p) >= 3 {
			policies = append(policies, &domain.Policy{
				Role:     p[0],
				Resource: p[1],
				Action:   p[2],
			})
		}
	}

	// Get available roles for the form dropdown
	roles, _ := h.roleService.List(r.Context())

	templates.PoliciesPage(policies, roles, auth).Render(r.Context(), w)
}

func (h *WebPolicyHandler) NewPolicyForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)
	roles, _ := h.roleService.List(r.Context())
	templates.PolicyForm(&domain.Policy{}, roles, "", auth).Render(r.Context(), w)
}

func (h *WebPolicyHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	if err := r.ParseForm(); err != nil {
		roles, _ := h.roleService.List(r.Context())
		templates.PolicyForm(&domain.Policy{}, roles, "Invalid form data", auth).Render(r.Context(), w)
		return
	}

	role := r.FormValue("role")
	resource := r.FormValue("resource")
	action := r.FormValue("action")

	if role == "" || resource == "" || action == "" {
		roles, _ := h.roleService.List(r.Context())
		templates.PolicyForm(&domain.Policy{Role: role, Resource: resource, Action: action}, roles, "All fields are required", auth).Render(r.Context(), w)
		return
	}

	added, err := h.enforcer.AddPolicy(role, resource, action)
	if err != nil {
		roles, _ := h.roleService.List(r.Context())
		templates.PolicyForm(&domain.Policy{Role: role, Resource: resource, Action: action}, roles, err.Error(), auth).Render(r.Context(), w)
		return
	}

	if !added {
		roles, _ := h.roleService.List(r.Context())
		templates.PolicyForm(&domain.Policy{Role: role, Resource: resource, Action: action}, roles, "Policy already exists", auth).Render(r.Context(), w)
		return
	}

	if err := h.enforcer.SavePolicy(); err != nil {
		roles, _ := h.roleService.List(r.Context())
		templates.PolicyForm(&domain.Policy{Role: role, Resource: resource, Action: action}, roles, "Failed to save policy: "+err.Error(), auth).Render(r.Context(), w)
		return
	}

	h.ListPolicies(w, r)
}

func (h *WebPolicyHandler) DeletePolicy(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	role := r.FormValue("role")
	resource := r.FormValue("resource")
	action := r.FormValue("action")

	removed, err := h.enforcer.RemovePolicy(role, resource, action)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !removed {
		http.Error(w, "Policy not found", http.StatusNotFound)
		return
	}

	if err := h.enforcer.SavePolicy(); err != nil {
		http.Error(w, "Failed to save policy: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
