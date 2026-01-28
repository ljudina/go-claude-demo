package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"claude-test/internal/domain"
	"claude-test/internal/handler/templates"
	"claude-test/internal/service"
)

type RoleWebHandler struct {
	roleService *service.RoleService
}

func NewRoleWebHandler(roleService *service.RoleService) *RoleWebHandler {
	return &RoleWebHandler{roleService: roleService}
}

func (h *RoleWebHandler) RegisterRoutes(r chi.Router) {
	r.Route("/web/roles", func(r chi.Router) {
		r.Get("/", h.ListRoles)
		r.Get("/new", h.NewRoleForm)
		r.Post("/", h.CreateRole)
		r.Get("/{id}/edit", h.EditRoleForm)
		r.Put("/{id}", h.UpdateRole)
		r.Delete("/{id}", h.DeleteRole)
	})
}

func (h *RoleWebHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.RolesPage(roles).Render(r.Context(), w)
}

func (h *RoleWebHandler) NewRoleForm(w http.ResponseWriter, r *http.Request) {
	templates.RoleForm(&domain.Role{}, false, "").Render(r.Context(), w)
}

func (h *RoleWebHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		templates.RoleForm(&domain.Role{}, false, "Invalid form data").Render(r.Context(), w)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	_, err := h.roleService.Create(r.Context(), name, description)
	if err != nil {
		templates.RoleForm(&domain.Role{Name: name, Description: description}, false, err.Error()).Render(r.Context(), w)
		return
	}

	h.ListRoles(w, r)
}

func (h *RoleWebHandler) EditRoleForm(w http.ResponseWriter, r *http.Request) {
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

	templates.RoleForm(role, true, "").Render(r.Context(), w)
}

func (h *RoleWebHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		templates.RoleForm(&domain.Role{ID: id}, true, "Invalid form data").Render(r.Context(), w)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")

	_, err = h.roleService.Update(r.Context(), id, name, description)
	if err != nil {
		templates.RoleForm(&domain.Role{ID: id, Name: name, Description: description}, true, err.Error()).Render(r.Context(), w)
		return
	}

	h.ListRoles(w, r)
}

func (h *RoleWebHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
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
