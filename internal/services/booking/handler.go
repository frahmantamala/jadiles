package booking

import (
	"net/http"
	"strconv"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

// Handler handles HTTP requests for booking capability
type Handler struct {
	service *ServiceUsecase
}

// NewHandler creates a new booking handler
func NewHandler(service *ServiceUsecase) *Handler {
	return &Handler{
		service: service,
	}
}

// CreateBooking handles POST /bookings
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated parent ID from JWT context
	parentID, err := internal.ExtractParentID(ctx)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Authentication required"))
		return
	}

	// Parse and validate request
	params, err := NewCreateBookingParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	if err := params.Validate(ctx); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Convert to domain request
	req, err := params.ToCreateBookingRequest(parentID)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError(err.Error()))
		return
	}

	// Create booking
	confirmation, err := h.service.CreateBooking(ctx, req)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Convert to v1 response
	response := ToV1BookingConfirmation(confirmation)

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, response)
}

// GetBooking handles GET /bookings/{booking_id}
func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authenticated parent ID from JWT context
	parentID, err := internal.ExtractParentID(ctx)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Authentication required"))
		return
	}

	// Parse booking ID from URL
	bookingIDStr := chi.URLParam(r, "booking_id")
	bookingID, err := strconv.ParseInt(bookingIDStr, 10, 64)
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewValidationError("booking_id must be a valid integer"))
		return
	}

	// Get booking
	response, err := h.service.GetBooking(ctx, bookingID, parentID)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
