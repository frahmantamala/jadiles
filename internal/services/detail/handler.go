package detail

import (
	"net/http"
	"strconv"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Handler handles HTTP requests for detail capability
type Handler struct {
	service *ServiceUsecase
}

// NewHandler creates a new detail handler
func NewHandler(service *ServiceUsecase) *Handler {
	return &Handler{
		service: service,
	}
}

// GetServiceDetail handles GET /services/{service_id}
func (h *Handler) GetServiceDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse service ID from URL
	serviceIDStr := chi.URLParam(r, "service_id")
	serviceID, err := strconv.ParseInt(serviceIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("service_id must be a valid integer"))
		return
	}

	// Get service detail
	response, err := h.service.GetServiceDetail(ctx, serviceID)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
