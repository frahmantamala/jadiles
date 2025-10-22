// this is for using handler http
package auth

import (
	"encoding/json"
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/render"
)

// Handler handles HTTP requests for authentication
type Handler struct {
	service *Service
}

// NewHandler creates a new auth handler
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterParent handles parent registration
func (h *Handler) RegisterParent(w http.ResponseWriter, r *http.Request) {
	var req v1.RegisterParentRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.HandleEndpointError(w, r, internal.NewAppError("INVALID_JSON", "Invalid request body", http.StatusBadRequest, err))
		return
	}

	// Call service
	resp, err := h.service.RegisterParent(r.Context(), req)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Return response
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// RegisterVendor handles vendor registration
func (h *Handler) RegisterVendor(w http.ResponseWriter, r *http.Request) {
	var req v1.RegisterVendorRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.HandleEndpointError(w, r, internal.NewAppError("INVALID_JSON", "Invalid request body", http.StatusBadRequest, err))
		return
	}

	// Call service
	resp, err := h.service.RegisterVendor(r.Context(), req)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Return response
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req v1.LoginRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		internal.HandleEndpointError(w, r, internal.NewAppError("INVALID_JSON", "Invalid request body", http.StatusBadRequest, err))
		return
	}

	// Call service
	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// Logout handles user logout (token invalidation)
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID and token from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	token, ok := r.Context().Value("token").(string)
	if !ok {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Call service
	if err := h.service.Logout(r.Context(), userID, token); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Logout successful",
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
