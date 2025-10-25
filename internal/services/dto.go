package services

import (
	"context"
	"net/http"
	"strconv"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/core/common"
	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// SearchServicesParams represents query parameters for searching services
type SearchServicesParams struct {
	CategorySlug string      `validate:"omitempty"`
	City         string      `validate:"omitempty,max=100"`
	District     string      `validate:"omitempty,max=100"`
	ChildAge     *int        `validate:"omitempty,min=0,max=18"`
	SkillLevel   *SkillLevel `validate:"omitempty,oneof=beginner intermediate advanced all_levels"`
	ClassType    *ClassType  `validate:"omitempty,oneof=private small_group large_group"`
	DayOfWeek    *int        `validate:"omitempty,min=0,max=6"`
	MinPrice     *float64    `validate:"omitempty,min=0"`
	MaxPrice     *float64    `validate:"omitempty,min=0"`
	MinRating    *float64    `validate:"omitempty,min=0,max=5"`
	FeaturedOnly bool        `validate:"omitempty"`
	Page         int         `validate:"required,min=1"`
	Limit        int         `validate:"required,min=1,max=100"`
}

// NewSearchServicesParams creates SearchServicesParams from HTTP request query parameters
func NewSearchServicesParams(r *http.Request) (*SearchServicesParams, error) {
	query := r.URL.Query()

	params := &SearchServicesParams{
		CategorySlug: query.Get("category"),
		City:         query.Get("city"),
		District:     query.Get("district"),
		Page:         1,  // Default
		Limit:        20, // Default
	}

	// Parse page
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, internal.NewValidationError("page must be a valid integer")
		}
		params.Page = page
	}

	// Parse limit
	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, internal.NewValidationError("limit must be a valid integer")
		}
		params.Limit = limit
	}

	// Parse age
	if ageStr := query.Get("age"); ageStr != "" {
		age, err := strconv.Atoi(ageStr)
		if err != nil {
			return nil, internal.NewValidationError("age must be a valid integer")
		}
		params.ChildAge = &age
	}

	// Parse skill level
	if skillStr := query.Get("skill_level"); skillStr != "" {
		skill := SkillLevel(skillStr)
		params.SkillLevel = &skill
	}

	// Parse class type
	if classStr := query.Get("class_type"); classStr != "" {
		class := ClassType(classStr)
		params.ClassType = &class
	}

	// Parse day of week (convert string to int)
	if dayStr := query.Get("day"); dayStr != "" {
		dayOfWeek, err := parseDayOfWeek(dayStr)
		if err != nil {
			return nil, err
		}
		params.DayOfWeek = &dayOfWeek
	}

	// Parse min price
	if minPriceStr := query.Get("price_min"); minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			return nil, internal.NewValidationError("price_min must be a valid number")
		}
		params.MinPrice = &minPrice
	}

	// Parse max price
	if maxPriceStr := query.Get("price_max"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			return nil, internal.NewValidationError("price_max must be a valid number")
		}
		params.MaxPrice = &maxPrice
	}

	// Parse min rating
	if ratingStr := query.Get("rating_min"); ratingStr != "" {
		minRating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			return nil, internal.NewValidationError("rating_min must be a valid number")
		}
		params.MinRating = &minRating
	}

	// Parse featured flag
	if featuredStr := query.Get("is_featured"); featuredStr != "" {
		featured, err := strconv.ParseBool(featuredStr)
		if err != nil {
			return nil, internal.NewValidationError("is_featured must be a valid boolean")
		}
		params.FeaturedOnly = featured
	}

	return params, nil
}

// parseDayOfWeek converts day name string to integer (0-6)
func parseDayOfWeek(dayStr string) (int, error) {
	dayMap := map[string]int{
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		"sunday":    0,
	}

	day, ok := dayMap[dayStr]
	if !ok {
		return 0, internal.NewValidationError("day must be one of: monday, tuesday, wednesday, thursday, friday, saturday, sunday")
	}

	return day, nil
}

// Validate validates SearchServicesParams
func (p *SearchServicesParams) Validate(ctx context.Context) error {
	if err := common.ValidateStruct(p); err != nil {
		return err
	}

	// Validate price range
	if p.MinPrice != nil && p.MaxPrice != nil && *p.MinPrice > *p.MaxPrice {
		return internal.NewValidationError("price_min cannot be greater than price_max")
	}

	// Validate skill level
	if p.SkillLevel != nil {
		if !p.SkillLevel.IsValid() {
			return internal.NewValidationError("invalid skill_level")
		}
	}

	// Validate class type
	if p.ClassType != nil {
		if !p.ClassType.IsValid() {
			return internal.NewValidationError("invalid class_type")
		}
	}

	return nil
}

// ToSearchFilters converts SearchServicesParams to SearchFilters
func (p *SearchServicesParams) ToSearchFilters() *SearchFilters {
	filters := &SearchFilters{
		CategorySlug: p.CategorySlug,
		City:         p.City,
		District:     p.District,
		ChildAge:     p.ChildAge,
		SkillLevel:   p.SkillLevel,
		ClassType:    p.ClassType,
		DayOfWeek:    p.DayOfWeek,
		MinPrice:     p.MinPrice,
		MaxPrice:     p.MaxPrice,
		MinRating:    p.MinRating,
		FeaturedOnly: p.FeaturedOnly,
		Page:         p.Page,
		PageSize:     p.Limit,
		SortBy:       "featured_first", // Default sort
	}

	return filters
}

// ServiceSearchResult represents the search result with services and pagination
type ServiceSearchResult struct {
	Services   []*Service
	Pagination *Pagination
}

// Pagination holds pagination metadata
type Pagination struct {
	Page       int
	Limit      int
	Total      int
	TotalPages int
}

// ToV1SearchResponse converts ServiceSearchResult to v1.ServicesSearchResponse
func ToV1SearchResponse(result *ServiceSearchResult) *v1.ServicesSearchResponse {
	response := &v1.ServicesSearchResponse{}

	// Convert services
	services := make([]v1.ServiceWithVendor, 0, len(result.Services))
	for _, svc := range result.Services {
		services = append(services, ToV1ServiceWithVendor(svc))
	}

	// Build pagination
	page := result.Pagination.Page
	limit := result.Pagination.Limit
	total := result.Pagination.Total
	totalPages := result.Pagination.TotalPages

	// Build response structure
	response.Data.Services = &services
	response.Data.Pagination = &v1.Pagination{
		Page:       &page,
		Limit:      &limit,
		Total:      &total,
		TotalPages: &totalPages,
	}

	return response
}

// ToV1ServiceWithVendor converts domain Service to v1.ServiceWithVendor
func ToV1ServiceWithVendor(s *Service) v1.ServiceWithVendor {
	// Build base service fields
	id := s.ID
	name := s.Name
	description := s.Description
	ageMin := s.AgeMin
	ageMax := s.AgeMax
	ageRange := s.GetAgeRangeDisplay()
	skillLevel := v1.SkillLevel(s.SkillLevel)
	classType := v1.ClassType(s.ClassType)
	maxParticipants := s.MaxParticipants
	durationMinutes := s.DurationMinutes
	pricePerSession := s.PricePerSession
	isFeatured := s.IsFeatured
	status := v1.ServiceStatus(s.Status)

	service := v1.ServiceWithVendor{
		Id:              &id,
		Name:            &name,
		Description:     &description,
		AgeMin:          &ageMin,
		AgeMax:          &ageMax,
		AgeRange:        &ageRange,
		SkillLevel:      &skillLevel,
		ClassType:       &classType,
		MaxParticipants: &maxParticipants,
		DurationMinutes: &durationMinutes,
		PricePerSession: &pricePerSession,
		IsFeatured:      &isFeatured,
		Status:          &status,
	}

	// Add optional pricing fields
	if s.TrialPrice != nil {
		service.TrialPrice = s.TrialPrice
	}
	if s.Package4Price != nil {
		service.Package4Price = s.Package4Price
	}
	if s.Package8Price != nil {
		service.Package8Price = s.Package8Price
	}
	if s.Package12Price != nil {
		service.Package12Price = s.Package12Price
	}
	if s.Requirements != nil {
		service.Requirements = s.Requirements
	}
	if s.WhatWillLearn != nil {
		service.WhatWillLearn = s.WhatWillLearn
	}

	// Add vendor info
	if s.VendorID != 0 {
		vendorID := s.VendorID
		businessName := s.VendorBusinessName
		city := s.VendorCity
		district := s.VendorDistrict
		verified := s.VendorVerified

		vendor := v1.Vendor{
			Id:           &vendorID,
			BusinessName: &businessName,
			City:         &city,
			District:     &district,
			Verified:     &verified,
		}

		if s.VendorLogo != nil {
			vendor.Logo = s.VendorLogo
		}
		if s.VendorRatingAvg != nil {
			vendor.RatingAvg = s.VendorRatingAvg
		}
		if s.VendorTotalReviews != nil {
			vendor.TotalReviews = s.VendorTotalReviews
		}

		service.Vendor = &vendor
	}

	// Add category info
	if s.CategoryID != 0 {
		categoryID := s.CategoryID
		categoryName := s.CategoryName
		categorySlug := s.CategorySlug

		category := v1.Category{
			Id:   &categoryID,
			Name: &categoryName,
			Slug: &categorySlug,
		}

		service.Category = &category
	}

	// Add available days
	if len(s.AvailableDays) > 0 {
		availableDays := make([]string, len(s.AvailableDays))
		for i, day := range s.AvailableDays {
			availableDays[i] = day
		}
		service.AvailableDays = &availableDays
	}

	// Add next available date
	if s.NextAvailable != nil {
		nextAvail := openapi_types.Date{Time: *s.NextAvailable}
		service.NextAvailable = &nextAvail
	}

	// Add distance if available
	if s.DistanceKm != nil {
		service.DistanceKm = s.DistanceKm
	}

	return service
}

// ToV1Category converts datamodel.ServiceCategory to v1.Category
func ToV1Category(c *datamodel.ServiceCategory) v1.Category {
	id := c.ID
	name := c.Name
	slug := c.Slug
	displayOrder := int(c.DisplayOrder)

	category := v1.Category{
		Id:           &id,
		Name:         &name,
		Slug:         &slug,
		Description:  c.Description, // Already a pointer
		DisplayOrder: &displayOrder,
	}

	if c.Icon != nil {
		category.Icon = c.Icon
	}
	if c.ColorHex != nil {
		category.ColorHex = c.ColorHex
	}

	return category
}
