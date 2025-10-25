package detail

import (
	"context"
	"database/sql"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// Repository defines the data access interface for detail capability
type Repository interface {
	GetServiceDetail(ctx context.Context, serviceID int64) (*postgresql.ServiceDetailData, error)
}

// ServiceUsecase handles service detail business logic
type ServiceUsecase struct {
	repo Repository
}

// NewService creates a new detail service
func NewService(repo Repository) *ServiceUsecase {
	return &ServiceUsecase{
		repo: repo,
	}
}

// GetServiceDetail retrieves comprehensive service information
func (s *ServiceUsecase) GetServiceDetail(ctx context.Context, serviceID int64) (*v1.ServiceDetailResponse, error) {
	// Fetch all service detail data
	data, err := s.repo.GetServiceDetail(ctx, serviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, internal.NewNotFoundError("Service not found")
		}
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to domain ServiceDetail
	serviceDetail, err := postgresql.ToServiceDetail(data)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to v1.ServiceDetailResponse
	response := ToV1ServiceDetailResponse(serviceDetail)

	return response, nil
}
