package schedule

import (
	"net/http"
	"strconv"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Handler handles HTTP requests for schedule capability
type Handler struct {
	service *ServiceUsecase
}

// NewHandler creates a new schedule handler
func NewHandler(service *ServiceUsecase) *Handler {
	return &Handler{
		service: service,
	}
}

// GetServiceAvailability handles GET /services/{service_id}/availability
func (h *Handler) GetServiceAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse service ID from URL
	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("service_id must be a valid integer"))
		return
	}

	// Parse query parameters
	params, err := NewGetAvailabilityParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Validate params
	if err := params.Validate(ctx); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Get availability
	response, err := h.service.GetServiceAvailability(ctx, serviceID, params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
