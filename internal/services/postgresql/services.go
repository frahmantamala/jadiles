package postgresql

import (
	"context"
	"fmt"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	"github.com/frahmantamala/jadiles/internal/services"
	"gorm.io/gorm"
)

// searchResult represents the raw query result with embedded fields and aggregated data
type searchResult struct {
	datamodel.Services
	VendorBusinessName string   `gorm:"column:vendor_business_name"`
	VendorCity         string   `gorm:"column:vendor_city"`
	VendorDistrict     string   `gorm:"column:vendor_district"`
	VendorLogo         *string  `gorm:"column:vendor_logo"`
	VendorRatingAvg    *float64 `gorm:"column:vendor_rating_avg"`
	VendorTotalReviews *int     `gorm:"column:vendor_total_reviews"`
	VendorVerified     bool     `gorm:"column:vendor_verified"`
	CategoryName       string   `gorm:"column:category_name"`
	CategorySlug       string   `gorm:"column:category_slug"`
}

// SearchServices performs complex filtering and searching for services
func (r *Repository) SearchServices(ctx context.Context, filters *services.SearchFilters) ([]*services.ServiceWithAggregates, int64, error) {
	var servicesData []*searchResult
	var total int64

	// Build base query with joins
	query := r.db.WithContext(ctx).
		Table("services s").
		Select(`
			s.id,
			s.vendor_id,
			s.category_id,
			s.name,
			s.description,
			s.age_min,
			s.age_max,
			s.skill_level,
			s.class_type,
			s.max_participants,
			s.duration_minutes,
			s.price_per_session,
			s.trial_price,
			s.package_4_price,
			s.package_8_price,
			s.package_12_price,
			s.requirements,
			s.what_will_learn,
			s.is_featured,
			s.status,
			s.created_at,
			s.updated_at,
			v.business_name as vendor_business_name,
			v.city as vendor_city,
			v.district as vendor_district,
			v.logo as vendor_logo,
			v.rating_avg as vendor_rating_avg,
			v.total_reviews as vendor_total_reviews,
			v.verified as vendor_verified,
			sc.name as category_name,
			sc.slug as category_slug
		`).
		Joins("INNER JOIN vendors v ON s.vendor_id = v.id").
		Joins("INNER JOIN service_categories sc ON s.category_id = sc.id").
		Where("s.status = ?", "active").
		Where("v.status = ?", "approved")

	// Apply filters
	query = r.applyFilters(query, filters)

	// Count total records (before pagination)
	countQuery := r.db.WithContext(ctx).
		Table("services s").
		Joins("INNER JOIN vendors v ON s.vendor_id = v.id").
		Joins("INNER JOIN service_categories sc ON s.category_id = sc.id").
		Where("s.status = ?", "active").
		Where("v.status = ?", "approved")
	countQuery = r.applyFilters(countQuery, filters)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	query = r.applySorting(query, filters.SortBy)

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	query = query.Limit(filters.PageSize).Offset(offset)

	// Execute query
	if err := query.Scan(&servicesData).Error; err != nil {
		return nil, 0, err
	}

	// Convert to ServiceWithAggregates
	result := make([]*services.ServiceWithAggregates, len(servicesData))
	for i, data := range servicesData {
		result[i] = &services.ServiceWithAggregates{
			Services:           data.Services,
			VendorBusinessName: data.VendorBusinessName,
			VendorCity:         data.VendorCity,
			VendorDistrict:     data.VendorDistrict,
			VendorLogo:         data.VendorLogo,
			VendorRatingAvg:    data.VendorRatingAvg,
			VendorTotalReviews: data.VendorTotalReviews,
			VendorVerified:     data.VendorVerified,
			CategoryName:       data.CategoryName,
			CategorySlug:       data.CategorySlug,
		}
	}

	return result, total, nil
}

// applyFilters applies all search filters to the query
func (r *Repository) applyFilters(query *gorm.DB, filters *services.SearchFilters) *gorm.DB {
	// Filter by category
	if filters.CategoryID != nil {
		query = query.Where("s.category_id = ?", *filters.CategoryID)
	}
	if filters.CategorySlug != "" {
		query = query.Where("sc.slug = ?", filters.CategorySlug)
	}

	// Filter by location
	if filters.City != "" {
		query = query.Where("v.city = ?", filters.City)
	}
	if filters.District != "" {
		query = query.Where("v.district = ?", filters.District)
	}

	// Filter by child age
	if filters.ChildAge != nil {
		query = query.Where("s.age_min <= ?", *filters.ChildAge).
			Where("s.age_max >= ?", *filters.ChildAge)
	}

	// Filter by skill level
	if filters.SkillLevel != nil {
		if *filters.SkillLevel == services.SkillLevelAllLevels {
			query = query.Where("s.skill_level = ?", "all_levels")
		} else {
			query = query.Where("(s.skill_level = ? OR s.skill_level = ?)",
				*filters.SkillLevel, services.SkillLevelAllLevels)
		}
	}

	// Filter by class type
	if filters.ClassType != nil {
		query = query.Where("s.class_type = ?", *filters.ClassType)
	}

	// Filter by day of week (requires schedule join)
	if filters.DayOfWeek != nil {
		query = query.Joins("INNER JOIN schedules sch ON s.id = sch.service_id").
			Where("sch.day_of_week = ?", *filters.DayOfWeek).
			Distinct()
	}

	// Filter by price range
	if filters.MinPrice != nil {
		query = query.Where("s.price_per_session >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("s.price_per_session <= ?", *filters.MaxPrice)
	}

	// Filter by minimum rating
	if filters.MinRating != nil {
		query = query.Where("v.rating_avg >= ?", *filters.MinRating)
	}

	// Filter by featured
	if filters.FeaturedOnly {
		query = query.Where("s.is_featured = ?", true)
	}

	return query
}

// applySorting applies sorting to the query
func (r *Repository) applySorting(query *gorm.DB, sortBy string) *gorm.DB {
	switch sortBy {
	case "price_asc":
		return query.Order("s.price_per_session ASC")
	case "price_desc":
		return query.Order("s.price_per_session DESC")
	case "rating":
		return query.Order("v.rating_avg DESC NULLS LAST")
	case "newest":
		return query.Order("s.created_at DESC")
	case "featured_first":
		fallthrough
	default:
		// Featured first, then by rating, then by created date
		return query.Order("s.is_featured DESC, v.rating_avg DESC NULLS LAST, s.created_at DESC")
	}
}

// GetServiceByID retrieves a service by ID
func (r *Repository) GetServiceByID(ctx context.Context, id int64) (*datamodel.Services, error) {
	var service datamodel.Services
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&service).Error

	if err != nil {
		return nil, err
	}

	return &service, nil
}

// GetCategoryBySlug retrieves a category by slug
func (r *Repository) GetCategoryBySlug(ctx context.Context, slug string) (*datamodel.ServiceCategory, error) {
	var category datamodel.ServiceCategory
	err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&category).Error

	if err != nil {
		return nil, err
	}

	return &category, nil
}

// GetAvailableDaysForService retrieves available days for a service
func (r *Repository) GetAvailableDaysForService(ctx context.Context, serviceID int64) ([]string, error) {
	var schedules []struct {
		DayOfWeek int
	}

	err := r.db.WithContext(ctx).
		Table("schedules").
		Select("DISTINCT day_of_week").
		Where("service_id = ?", serviceID).
		Order("day_of_week ASC").
		Scan(&schedules).Error

	if err != nil {
		return nil, err
	}

	// Convert day numbers to names
	dayNames := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	result := make([]string, 0, len(schedules))
	for _, s := range schedules {
		if s.DayOfWeek >= 0 && s.DayOfWeek < len(dayNames) {
			result = append(result, dayNames[s.DayOfWeek])
		}
	}

	return result, nil
}

// GetNextAvailableDate retrieves the next available date for a service
func (r *Repository) GetNextAvailableDate(ctx context.Context, serviceID int64) (*string, error) {
	var result struct {
		NextDate string
	}

	// Query to find the next available date from schedules
	// This is a simplified version - in production you'd want to consider:
	// - Existing bookings
	// - Available slots
	// - Schedule exceptions
	query := `
		SELECT
			TO_CHAR(
				CURRENT_DATE + ((sch.day_of_week - EXTRACT(DOW FROM CURRENT_DATE)::INTEGER + 7) % 7)::INTEGER,
				'YYYY-MM-DD'
			) as next_date
		FROM schedules sch
		WHERE sch.service_id = ?
			AND sch.available_slots > 0
		ORDER BY (sch.day_of_week - EXTRACT(DOW FROM CURRENT_DATE)::INTEGER + 7) % 7
		LIMIT 1
	`

	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&result).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	if result.NextDate == "" {
		return nil, nil
	}

	return &result.NextDate, nil
}

// GetAllCategories retrieves all service categories
func (r *Repository) GetAllCategories(ctx context.Context) ([]*datamodel.ServiceCategory, error) {
	var categories []*datamodel.ServiceCategory
	err := r.db.WithContext(ctx).
		Order("display_order ASC, name ASC").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// EnrichServiceWithDetails enriches service data with aggregated information
func (r *Repository) EnrichServiceWithDetails(ctx context.Context, serviceID int64) (map[string]interface{}, error) {
	details := make(map[string]interface{})

	// Get available days
	days, err := r.GetAvailableDaysForService(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available days: %w", err)
	}
	details["available_days"] = days

	// Get next available date
	nextDate, err := r.GetNextAvailableDate(ctx, serviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next available date: %w", err)
	}
	details["next_available"] = nextDate

	return details, nil
}
