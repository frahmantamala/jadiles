package review

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/services"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// GetReviewsParams represents query parameters for reviews endpoint
type GetReviewsParams struct {
	Page  int `validate:"required,min=1"`
	Limit int `validate:"required,min=1,max=100"`
}

// NewGetReviewsParams creates GetReviewsParams from HTTP request
func NewGetReviewsParams(r *http.Request) (*GetReviewsParams, error) {
	params := &GetReviewsParams{
		Page:  1,  // Default
		Limit: 10, // Default
	}

	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, internal.NewValidationError("page must be a valid integer")
		}
		params.Page = page
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, internal.NewValidationError("limit must be a valid integer")
		}
		params.Limit = limit
	}

	return params, nil
}

// Validate validates GetReviewsParams
func (p *GetReviewsParams) Validate(ctx context.Context) error {
	return common.ValidateStruct(p)
}

// ReviewsResult represents paginated reviews with summary
type ReviewsResult struct {
	Reviews    []*services.ReviewPreview
	Summary    *services.ReviewSummary
	Pagination *ReviewPagination
}

// ReviewPagination holds pagination metadata for reviews
type ReviewPagination struct {
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

// ToV1ReviewsResponse converts domain reviews to v1.ServiceReviewsResponse
func ToV1ReviewsResponse(result *ReviewsResult) *v1.ServiceReviewsResponse {
	// TODO: Implement proper conversion when OpenAPI schema is defined with review types
	// For now, return empty response to allow compilation
	_ = result
	return &v1.ServiceReviewsResponse{}
}

// ToReviewPreviews converts postgresql review data to domain models
func ToReviewPreviews(data []*postgresql.ReviewPreviewData) ([]*services.ReviewPreview, error) {
	reviews := make([]*services.ReviewPreview, 0, len(data))
	for _, r := range data {
		var photos []string
		// Parse JSONB photos if needed - skipped for now

		var respondedAt *time.Time
		if r.RespondedAt != nil && r.RespondedAt.Valid {
			t := r.RespondedAt.Time
			respondedAt = &t
		}

		var createdAt time.Time
		if r.CreatedAt.Valid {
			createdAt = r.CreatedAt.Time
		}

		review := &services.ReviewPreview{
			ID:             r.ID,
			ParentName:     r.ParentName,
			ChildAge:       r.ChildAge,
			Rating:         r.Rating,
			ReviewText:     r.ReviewText,
			DidChildEnjoy:  r.DidChildEnjoy,
			WouldRecommend: r.WouldRecommend,
			Photos:         photos,
			VendorResponse: r.VendorResponse,
			RespondedAt:    respondedAt,
			CreatedAt:      createdAt,
		}

		reviews = append(reviews, review)
	}
	return reviews, nil
}

// ToReviewSummary converts postgresql review summary data to domain model
func ToReviewSummary(data *postgresql.ReviewSummaryData) *services.ReviewSummary {
	return &services.ReviewSummary{
		TotalReviews:  data.TotalReviews,
		AverageRating: data.AverageRating,
		RatingDistribution: map[int]int{
			1: data.Rating1,
			2: data.Rating2,
			3: data.Rating3,
			4: data.Rating4,
			5: data.Rating5,
		},
		ChildEnjoyedPercentage:   data.ChildEnjoyedPct,
		WouldRecommendPercentage: data.WouldRecommendPct,
	}
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) *ReviewPagination {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &ReviewPagination{
		Page:       page,
		Limit:      limit,
		Total:      int(total),
		TotalPages: totalPages,
	}
}
