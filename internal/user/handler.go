package user

import (
	"context"
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/render"
)

type ServiceAPI interface {
	RegisterParent(ctx context.Context, params *RegisterParentParams) (*v1.RegisterResponse, error)
	RegisterVendor(ctx context.Context, params *RegisterVendorParams) (*v1.RegisterVendorResponse, error)
	Login(ctx context.Context, params *LoginParams) (*v1.LoginResponse, error)
	Logout(ctx context.Context, userID int64, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*v1.LoginResponse, error)
	GetUserByID(ctx context.Context, userID int64) (*User, error)
}

type Handler struct {
	service ServiceAPI
}

func NewHandler(service ServiceAPI) *Handler {
	return &Handler{service: service}
}

// RegisterParent handles parent registration
func (h *Handler) RegisterParent(w http.ResponseWriter, r *http.Request) {
	params, err := NewRegisterParentParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	if err := params.Validate(r.Context()); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	resp, err := h.service.RegisterParent(r.Context(), params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// RegisterVendor handles vendor registration
func (h *Handler) RegisterVendor(w http.ResponseWriter, r *http.Request) {
	params, err := NewRegisterVendorParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	if err := params.Validate(r.Context()); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	resp, err := h.service.RegisterVendor(r.Context(), params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, resp)
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	params, err := NewLoginParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	err = params.Validate(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	resp, err := h.service.Login(r.Context(), params)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// Logout handles user logout (requires authentication)
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Extract token from context
	token, ok := internal.ExtractToken(r.Context())
	if !ok {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	// Call service to logout
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

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	params, err := NewRefreshTokenParams(r)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	if err := params.Validate(r.Context()); err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	resp, err := h.service.RefreshToken(r.Context(), params.RefreshToken)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}

// GetProfile handles getting user profile (requires authentication)
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context
	userID, err := internal.ExtractUserID(r.Context())
	if err != nil {
		internal.HandleEndpointError(w, r, internal.NewUnauthorizedError("Unauthorized"))
		return
	}

	user, err := h.service.GetUserByID(r.Context(), userID)
	if err != nil {
		internal.HandleEndpointError(w, r, err)
		return
	}

	// Convert to response
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"full_name":  user.FullName,
			"phone":      user.Phone,
			"role":       user.Role,
			"status":     user.Status,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		},
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, resp)
}
