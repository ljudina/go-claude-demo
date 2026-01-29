package webhandler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"claude-test/internal/domain"
	"claude-test/internal/handler"
	"claude-test/internal/handler/templates"
	"claude-test/internal/service"
)

type WebRoleHandler struct {
	roleService    *service.RoleService
	navItemService *service.NavItemService
	authHandler    *handler.AuthHandler
}

func NewWebRoleHandler(roleService *service.RoleService, navItemService *service.NavItemService, authHandler *handler.AuthHandler) *WebRoleHandler {
	return &WebRoleHandler{
		roleService:    roleService,
		navItemService: navItemService,
		authHandler:    authHandler,
	}
}

func (h *WebRoleHandler) getAuthInfo(r *http.Request) *templates.AuthInfo {
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

func (h *WebRoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	roles, err := h.roleService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.RolesPage(roles, auth).Render(r.Context(), w)
}

func (h *WebRoleHandler) NewRoleForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)
	templates.RoleForm(&domain.Role{}, false, "", auth).Render(r.Context(), w)
}

func (h *WebRoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	if err := r.ParseForm(); err != nil {
		templates.RoleForm(&domain.Role{}, false, "Invalid form data", auth).Render(r.Context(), w)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	_, err := h.roleService.Create(r.Context(), name, description)
	if err != nil {
		templates.RoleForm(&domain.Role{Name: name, Description: description}, false, err.Error(), auth).Render(r.Context(), w)
		return
	}

	h.ListRoles(w, r)
}

func (h *WebRoleHandler) EditRoleForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	role, err := h.roleService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	templates.RoleForm(role, true, "", auth).Render(r.Context(), w)
}

func (h *WebRoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		templates.RoleForm(&domain.Role{ID: id}, true, "Invalid form data", auth).Render(r.Context(), w)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	_, err = h.roleService.Update(r.Context(), id, name, description)
	if err != nil {
		templates.RoleForm(&domain.Role{ID: id, Name: name, Description: description}, true, err.Error(), auth).Render(r.Context(), w)
		return
	}

	h.ListRoles(w, r)
}

func (h *WebRoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	if err := h.roleService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
