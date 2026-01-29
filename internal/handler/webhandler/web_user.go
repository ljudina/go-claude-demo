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

type WebUserHandler struct {
	userService *service.UserService
	roleService *service.RoleService
	authHandler *handler.AuthHandler
}

func NewWebUserHandler(userService *service.UserService, roleService *service.RoleService, authHandler *handler.AuthHandler) *WebUserHandler {
	return &WebUserHandler{
		userService: userService,
		roleService: roleService,
		authHandler: authHandler,
	}
}

func (h *WebUserHandler) getAuthInfo(r *http.Request) *templates.AuthInfo {
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

	return &templates.AuthInfo{
		User:    user,
		Roles:   roles,
		IsAdmin: isAdmin,
	}
}

func (h *WebUserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	users, err := h.userService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	usersWithRoles := make([]templates.UserWithRoles, len(users))
	for i, user := range users {
		roles, _ := h.roleService.GetUserRoles(r.Context(), user.ID)
		usersWithRoles[i] = templates.UserWithRoles{
			User:  user,
			Roles: roles,
		}
	}

	templates.UsersPage(usersWithRoles, auth).Render(r.Context(), w)
}

func (h *WebUserHandler) NewUserForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)
	allRoles, _ := h.roleService.List(r.Context())
	templates.UserForm(&domain.User{}, nil, allRoles, false, "", auth).Render(r.Context(), w)
}

func (h *WebUserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	if err := r.ParseForm(); err != nil {
		allRoles, _ := h.roleService.List(r.Context())
		templates.UserForm(&domain.User{}, nil, allRoles, false, "Invalid form data", auth).Render(r.Context(), w)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	roleIDs := parseRoleIDs(r.Form["roles"])

	user, err := h.userService.Create(r.Context(), email, name)
	if err != nil {
		allRoles, _ := h.roleService.List(r.Context())
		templates.UserForm(&domain.User{Email: email, Name: name}, roleIDs, allRoles, false, err.Error(), auth).Render(r.Context(), w)
		return
	}

	if len(roleIDs) > 0 {
		h.roleService.SetUserRoles(r.Context(), user.ID, roleIDs)
	}

	h.ListUsers(w, r)
}

func (h *WebUserHandler) EditUserForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userRoles, _ := h.roleService.GetUserRoles(r.Context(), id)
	userRoleIDs := make([]int64, len(userRoles))
	for i, role := range userRoles {
		userRoleIDs[i] = role.ID
	}

	allRoles, _ := h.roleService.List(r.Context())
	templates.UserForm(user, userRoleIDs, allRoles, true, "", auth).Render(r.Context(), w)
}

func (h *WebUserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		allRoles, _ := h.roleService.List(r.Context())
		templates.UserForm(&domain.User{ID: id}, nil, allRoles, true, "Invalid form data", auth).Render(r.Context(), w)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	roleIDs := parseRoleIDs(r.Form["roles"])

	_, err = h.userService.Update(r.Context(), id, email, name)
	if err != nil {
		allRoles, _ := h.roleService.List(r.Context())
		templates.UserForm(&domain.User{ID: id, Email: email, Name: name}, roleIDs, allRoles, true, err.Error(), auth).Render(r.Context(), w)
		return
	}

	h.roleService.SetUserRoles(r.Context(), id, roleIDs)

	h.ListUsers(w, r)
}

func (h *WebUserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.userService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseRoleIDs(values []string) []int64 {
	var roleIDs []int64
	for _, v := range values {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			roleIDs = append(roleIDs, id)
		}
	}
	return roleIDs
}
