package review

import (
	"context"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// Repository defines the data access interface for review capability
type Repository interface {
	GetServiceReviews(ctx context.Context, serviceID int64, page, limit int) ([]*postgresql.ReviewPreviewData, int64, error)
	GetReviewSummary(ctx context.Context, serviceID int64) (*postgresql.ReviewSummaryData, error)
}

// ServiceUsecase handles review business logic
type ServiceUsecase struct {
	repo Repository
}

// NewService creates a new review service
func NewService(repo Repository) *ServiceUsecase {
	return &ServiceUsecase{
		repo: repo,
	}
}

// GetServiceReviews retrieves paginated reviews
func (s *ServiceUsecase) GetServiceReviews(ctx context.Context, serviceID int64, params *GetReviewsParams) (*v1.ServiceReviewsResponse, error) {
	// Fetch reviews
	reviewsData, total, err := s.repo.GetServiceReviews(ctx, serviceID, params.Page, params.Limit)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Fetch review summary
	summaryData, err := s.repo.GetReviewSummary(ctx, serviceID)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to domain models
	reviews, err := ToReviewPreviews(reviewsData)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	summary := ToReviewSummary(summaryData)
	pagination := CalculatePagination(params.Page, params.Limit, total)

	// Build result
	result := &ReviewsResult{
		Reviews:    reviews,
		Summary:    summary,
		Pagination: pagination,
	}

	// Convert to v1.ServiceReviewsResponse
	response := ToV1ReviewsResponse(result)

	return response, nil
}
