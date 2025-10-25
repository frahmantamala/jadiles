package child

import (
	"context"
	"net/http"
	"strconv"

	"github.com/frahmantamala/jadiles/internal"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type ServiceAPI interface {
	AddChild(ctx context.Context, parentID int64, params *AddChildParams) (*v1.ChildResponse, error)
	GetChildren(ctx context.Context, parentID int64) (*v1.ChildrenListResponse, error)
	GetChild(ctx context.Context, childID int64, parentID int64) (*v1.ChildResponse, error)
	UpdateChild(ctx context.Context, childID int64, parentID int64, params *UpdateChildParams) (*v1.ChildResponse, error)
	DeleteChild(ctx context.Context, childID int64, parentID int64) error
}

type Handler struct {
	service ServiceAPI
}

func NewHandler(service ServiceAPI) *Handler {
	return &Handler{service: service}
}

// AddChild handles adding a new child
func (h *Handler) AddChild(w http.ResponseWriter, r *http.Request) {
	// Extract parent ID from context (set by auth middleware)
	parentID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Parse request
	params, err := NewAddChildParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Validate request
	if err := params.Validate(r.Context()); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Add child
	resp, err := h.service.AddChild(r.Context(), parentID, params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// GetChildren handles getting all children for a parent
func (h *Handler) GetChildren(w http.ResponseWriter, r *http.Request) {
	// Extract parent ID from context
	parentID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Get children
	resp, err := h.service.GetChildren(r.Context(), parentID)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetChild handles getting a specific child
func (h *Handler) GetChild(w http.ResponseWriter, r *http.Request) {
	// Extract parent ID from context
	parentID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Extract child ID from URL
	childIDStr := chi.URLParam(r, "id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("Invalid child ID"))
		return
	}

	// Get child
	resp, err := h.service.GetChild(r.Context(), childID, parentID)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// UpdateChild handles updating a child
func (h *Handler) UpdateChild(w http.ResponseWriter, r *http.Request) {
	// Extract parent ID from context
	parentID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Extract child ID from URL
	childIDStr := chi.URLParam(r, "id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("Invalid child ID"))
		return
	}

	// Parse request
	params, err := NewUpdateChildParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Validate request
	if err := params.Validate(r.Context()); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Update child
	resp, err := h.service.UpdateChild(r.Context(), childID, parentID, params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// DeleteChild handles deleting a child
func (h *Handler) DeleteChild(w http.ResponseWriter, r *http.Request) {
	// Extract parent ID from context
	parentID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Extract child ID from URL
	childIDStr := chi.URLParam(r, "id")
	childID, err := strconv.ParseInt(childIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("Invalid child ID"))
		return
	}

	// Delete child
	if err := h.service.DeleteChild(r.Context(), childID, parentID); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"message": "Child deleted successfully",
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
