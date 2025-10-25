package services

import (
	"context"
	"database/sql"
	"math"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
)

// ServiceWithAggregates is defined in postgresql package
// Re-export it for use in the interface
type ServiceWithAggregates = struct {
	datamodel.Services
	VendorBusinessName string
	VendorCity         string
	VendorDistrict     string
	VendorLogo         *string
	VendorRatingAvg    *float64
	VendorTotalReviews *int
	VendorVerified     bool
	CategoryName       string
	CategorySlug       string
}

type Repository interface {
	SearchServices(ctx context.Context, filters *SearchFilters) ([]*ServiceWithAggregates, int64, error)
	GetServiceByID(ctx context.Context, id int64) (*datamodel.Services, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*datamodel.ServiceCategory, error)
	GetAvailableDaysForService(ctx context.Context, serviceID int64) ([]string, error)
	GetNextAvailableDate(ctx context.Context, serviceID int64) (*string, error)
	GetAllCategories(ctx context.Context) ([]*datamodel.ServiceCategory, error)
	EnrichServiceWithDetails(ctx context.Context, serviceID int64) (map[string]interface{}, error)
}

type ServiceUsecase struct {
	repo Repository
}

func NewService(repo Repository) *ServiceUsecase {
	return &ServiceUsecase{
		repo: repo,
	}
}

// SearchServices searches for services with filters and returns paginated results
func (s *ServiceUsecase) SearchServices(ctx context.Context, params *SearchServicesParams) (*ServiceSearchResult, error) {
	// Convert params to filters
	filters := params.ToSearchFilters()

	// If category slug is provided, get category ID
	if filters.CategorySlug != "" {
		category, err := s.repo.GetCategoryBySlug(ctx, filters.CategorySlug)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, internal.NewNotFoundError("Category not found")
			}
			return nil, internal.NewInternalServerError(err)
		}
		filters.CategoryID = &category.ID
	}

	// Search services
	servicesData, total, err := s.repo.SearchServices(ctx, filters)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to domain models with enriched data
	services := make([]*Service, 0, len(servicesData))
	for _, svcData := range servicesData {
		// Convert to domain service
		domainService := FromServiceWithAggregates(svcData)

		// Enrich with additional data (available days, next available date)
		details, err := s.repo.EnrichServiceWithDetails(ctx, svcData.ID)
		if err != nil {
			// Log error but continue - enrichment is not critical
			// In production, you'd want proper logging here
		} else {
			if days, ok := details["available_days"].([]string); ok {
				domainService.AvailableDays = days
			}
			if nextDate, ok := details["next_available"].(*string); ok && nextDate != nil {
				// Parse to time.Time
				// For now, we'll keep it as string in the domain model
			}
		}

		services = append(services, domainService)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(filters.PageSize)))

	result := &ServiceSearchResult{
		Services: services,
		Pagination: &Pagination{
			Page:       filters.Page,
			Limit:      filters.PageSize,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}

	return result, nil
}

// GetCategories retrieves all service categories
func (s *ServiceUsecase) GetCategories(ctx context.Context) (*v1.CategoriesResponse, error) {
	categoriesData, err := s.repo.GetAllCategories(ctx)
	if err != nil {
		return nil, internal.NewInternalServerError(err)
	}

	// Convert to v1 response
	categories := make([]v1.Category, 0, len(categoriesData))
	for _, cat := range categoriesData {
		categories = append(categories, ToV1Category(cat))
	}

	response := &v1.CategoriesResponse{}
	response.Data = categories

	return response, nil
}
