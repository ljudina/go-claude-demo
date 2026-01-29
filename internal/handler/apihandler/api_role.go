package apihandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"claude-test/internal/domain"
	"claude-test/internal/service"
)

type ApiRoleHandler struct {
	service *service.RoleService
}

func NewApiRoleHandler(service *service.RoleService) *ApiRoleHandler {
	return &ApiRoleHandler{service: service}
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SetUserRolesRequest struct {
	RoleIDs []int64 `json:"role_ids"`
}

func (h *ApiRoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := h.service.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		handleRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toRoleResponse(role))
}

func (h *ApiRoleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		handleRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toRoleResponse(role))
}

func (h *ApiRoleHandler) List(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.List(r.Context())
	if err != nil {
		handleRoleError(w, err)
		return
	}

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = toRoleResponse(role)
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ApiRoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	role, err := h.service.Update(r.Context(), id, req.Name, req.Description)
	if err != nil {
		handleRoleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toRoleResponse(role))
}

func (h *ApiRoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		handleRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ApiRoleHandler) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	roles, err := h.service.GetUserRoles(r.Context(), userID)
	if err != nil {
		handleRoleError(w, err)
		return
	}

	response := make([]RoleResponse, len(roles))
	for i, role := range roles {
		response[i] = toRoleResponse(role)
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *ApiRoleHandler) SetUserRoles(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req SetUserRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.SetUserRoles(r.Context(), userID, req.RoleIDs); err != nil {
		handleRoleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toRoleResponse(r *domain.Role) RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
	}
}

func handleRoleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrRoleNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrRoleAlreadyExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidRoleName):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
