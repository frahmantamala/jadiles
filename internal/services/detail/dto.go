package detail

import (
	"github.com/frahmantamala/jadiles/internal/services"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// ToV1ServiceDetailResponse converts domain ServiceDetail to v1.ServiceDetailResponse
func ToV1ServiceDetailResponse(detail *services.ServiceDetail) *v1.ServiceDetailResponse {
	// TODO: Implement proper conversion when OpenAPI schema is defined with detail types
	// For now, return empty response to allow compilation
	_ = detail
	return &v1.ServiceDetailResponse{}
}
