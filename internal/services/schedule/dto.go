package schedule

import (
	"context"
	"net/http"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/services"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// GetAvailabilityParams represents query parameters for availability endpoint
type GetAvailabilityParams struct {
	Month string `validate:"required,len=7"` // YYYY-MM format
}

// NewGetAvailabilityParams creates GetAvailabilityParams from HTTP request
func NewGetAvailabilityParams(r *http.Request) (*GetAvailabilityParams, error) {
	month := r.URL.Query().Get("month")
	if month == "" {
		return nil, internal.NewValidationError("month parameter is required (format: YYYY-MM)")
	}
	return &GetAvailabilityParams{Month: month}, nil
}

// Validate validates GetAvailabilityParams
func (p *GetAvailabilityParams) Validate(ctx context.Context) error {
	// Validate YYYY-MM format
	_, err := time.Parse("2006-01", p.Month)
	if err != nil {
		return internal.NewValidationError("month must be in YYYY-MM format")
	}
	return common.ValidateStruct(p)
}

// ToV1AvailabilityResponse converts domain availability to v1.ServiceAvailabilityResponse
func ToV1AvailabilityResponse(availability []*services.DayAvailability) *v1.ServiceAvailabilityResponse {
	// TODO: Update this when OpenAPI schema is properly defined with availability types
	// For now, return empty response to allow compilation
	return &v1.ServiceAvailabilityResponse{}
}
