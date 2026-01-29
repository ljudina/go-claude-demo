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

type WebNavItemHandler struct {
	navItemService *service.NavItemService
	roleService    *service.RoleService
	authHandler    *handler.AuthHandler
}

func NewWebNavItemHandler(navItemService *service.NavItemService, roleService *service.RoleService, authHandler *handler.AuthHandler) *WebNavItemHandler {
	return &WebNavItemHandler{
		navItemService: navItemService,
		roleService:    roleService,
		authHandler:    authHandler,
	}
}

func (h *WebNavItemHandler) getAuthInfo(r *http.Request) *templates.AuthInfo {
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

func (h *WebNavItemHandler) ListNavItems(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	items, err := h.navItemService.ListWithRoles(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.NavItemsPage(items, auth).Render(r.Context(), w)
}

func (h *WebNavItemHandler) NewNavItemForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	roles, err := h.roleService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parentItems, err := h.navItemService.GetAllTopLevel(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	templates.NavItemForm(&domain.NavItem{}, false, "", roles, parentItems, auth).Render(r.Context(), w)
}

func (h *WebNavItemHandler) CreateNavItem(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	if err := r.ParseForm(); err != nil {
		h.renderFormWithError(w, r, &domain.NavItem{}, false, "Invalid form data", auth)
		return
	}

	name := r.FormValue("name")
	url := r.FormValue("url")
	icon := r.FormValue("icon")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	var parentID *int64
	if pid := r.FormValue("parent_id"); pid != "" && pid != "0" {
		p, err := strconv.ParseInt(pid, 10, 64)
		if err == nil {
			parentID = &p
		}
	}

	roleIDs := h.parseRoleIDs(r)

	_, err := h.navItemService.Create(r.Context(), name, url, icon, parentID, sortOrder, roleIDs)
	if err != nil {
		item := &domain.NavItem{Name: name, URL: url, Icon: icon, ParentID: parentID, SortOrder: sortOrder}
		h.renderFormWithError(w, r, item, false, err.Error(), auth)
		return
	}

	h.ListNavItems(w, r)
}

func (h *WebNavItemHandler) EditNavItemForm(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid nav item ID", http.StatusBadRequest)
		return
	}

	item, err := h.navItemService.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	roles, err := h.roleService.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parentItems, err := h.navItemService.GetAllTopLevel(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter out the current item from parent options to prevent self-referencing
	filteredParents := make([]*domain.NavItem, 0, len(parentItems))
	for _, p := range parentItems {
		if p.ID != id {
			filteredParents = append(filteredParents, p)
		}
	}

	templates.NavItemForm(item, true, "", roles, filteredParents, auth).Render(r.Context(), w)
}

func (h *WebNavItemHandler) UpdateNavItem(w http.ResponseWriter, r *http.Request) {
	auth := h.getAuthInfo(r)

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid nav item ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderFormWithError(w, r, &domain.NavItem{ID: id}, true, "Invalid form data", auth)
		return
	}

	name := r.FormValue("name")
	url := r.FormValue("url")
	icon := r.FormValue("icon")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	var parentID *int64
	if pid := r.FormValue("parent_id"); pid != "" && pid != "0" {
		p, err := strconv.ParseInt(pid, 10, 64)
		if err == nil {
			parentID = &p
		}
	}

	roleIDs := h.parseRoleIDs(r)

	_, err = h.navItemService.Update(r.Context(), id, name, url, icon, parentID, sortOrder, roleIDs)
	if err != nil {
		item := &domain.NavItem{ID: id, Name: name, URL: url, Icon: icon, ParentID: parentID, SortOrder: sortOrder}
		h.renderFormWithError(w, r, item, true, err.Error(), auth)
		return
	}

	h.ListNavItems(w, r)
}

func (h *WebNavItemHandler) DeleteNavItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid nav item ID", http.StatusBadRequest)
		return
	}

	if err := h.navItemService.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *WebNavItemHandler) parseRoleIDs(r *http.Request) []int64 {
	var roleIDs []int64
	for _, roleStr := range r.Form["roles"] {
		if roleID, err := strconv.ParseInt(roleStr, 10, 64); err == nil {
			roleIDs = append(roleIDs, roleID)
		}
	}
	return roleIDs
}

func (h *WebNavItemHandler) renderFormWithError(w http.ResponseWriter, r *http.Request, item *domain.NavItem, isEdit bool, errMsg string, auth *templates.AuthInfo) {
	roles, _ := h.roleService.List(r.Context())
	parentItems, _ := h.navItemService.GetAllTopLevel(r.Context())

	if isEdit && item.ID > 0 {
		// Filter out self from parents
		filtered := make([]*domain.NavItem, 0, len(parentItems))
		for _, p := range parentItems {
			if p.ID != item.ID {
				filtered = append(filtered, p)
			}
		}
		parentItems = filtered
	}

	templates.NavItemForm(item, isEdit, errMsg, roles, parentItems, auth).Render(r.Context(), w)
}
