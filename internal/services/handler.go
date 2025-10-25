package services

import (
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/go-chi/render"
)

type Handler struct {
	service *ServiceUsecase
}

func NewHandler(service *ServiceUsecase) *Handler {
	return &Handler{
		service: service,
	}
}

// SearchServices handles GET /services/search
func (h *Handler) SearchServices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	params, err := NewSearchServicesParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Validate params
	if err := params.Validate(ctx); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Search services
	result, err := h.service.SearchServices(ctx, params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Convert to v1 response
	response := ToV1SearchResponse(result)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// GetCategories handles GET /categories
func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get all categories
	response, err := h.service.GetCategories(ctx)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}
