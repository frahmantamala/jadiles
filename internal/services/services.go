package services

import (
	"fmt"
	"time"
)

// Service represents the service domain model
type Service struct {
	ID              int64
	VendorID        int64
	CategoryID      int64
	Name            string
	Description     string
	AgeMin          int
	AgeMax          int
	SkillLevel      SkillLevel
	ClassType       ClassType
	MaxParticipants int
	DurationMinutes int
	PricePerSession float64
	TrialPrice      *float64
	Package4Price   *float64
	Package8Price   *float64
	Package12Price  *float64
	Requirements    *string
	WhatWillLearn   *string
	Status          ServiceStatus
	IsFeatured      bool
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Aggregated data (not in services table)
	VendorBusinessName string
	VendorCity         string
	VendorDistrict     string
	VendorLogo         *string
	VendorRatingAvg    *float64
	VendorTotalReviews *int
	VendorVerified     bool
	CategoryName       string
	CategorySlug       string
	AvailableDays      []string
	NextAvailable      *time.Time
	DistanceKm         *float64
}

type SkillLevel string

const (
	SkillLevelBeginner     SkillLevel = "beginner"
	SkillLevelIntermediate SkillLevel = "intermediate"
	SkillLevelAdvanced     SkillLevel = "advanced"
	SkillLevelAllLevels    SkillLevel = "all_levels"
)

type ClassType string

const (
	ClassTypePrivate    ClassType = "private"
	ClassTypeSmallGroup ClassType = "small_group"
	ClassTypeLargeGroup ClassType = "large_group"
)

type ServiceStatus string

const (
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusInactive ServiceStatus = "inactive"
	ServiceStatusDraft    ServiceStatus = "draft"
)

// ServiceCategory represents a service category
type ServiceCategory struct {
	ID           int64
	Name         string
	Slug         string
	Description  string
	Icon         string
	ColorHex     string
	DisplayOrder int
	IsActive     bool
	CreatedAt    time.Time
}

// SearchFilters represents all possible search filters
type SearchFilters struct {
	// Category
	CategoryID   *int64
	CategorySlug string

	// Location
	City     string
	District string

	// Age filtering
	ChildAge *int

	// Service attributes
	SkillLevel *SkillLevel
	ClassType  *ClassType
	DayOfWeek  *int // 0=Sunday, 6=Saturday

	// Price range
	MinPrice *float64
	MaxPrice *float64

	// Rating
	MinRating *float64

	// Featured
	FeaturedOnly bool

	// Pagination
	Page     int
	PageSize int

	// Sorting
	SortBy string // price_asc, price_desc, rating, newest, featured
}

// Validate validates service fields
func (s *Service) Validate() error {
	if s.VendorID == 0 {
		return fmt.Errorf("vendor_id is required")
	}

	if s.CategoryID == 0 {
		return fmt.Errorf("category_id is required")
	}

	if s.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(s.Name) < 3 || len(s.Name) > 200 {
		return fmt.Errorf("name must be between 3 and 200 characters")
	}

	if s.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(s.Description) < 20 {
		return fmt.Errorf("description must be at least 20 characters")
	}

	if s.DurationMinutes <= 0 {
		return fmt.Errorf("duration must be greater than 0")
	}

	if s.PricePerSession <= 0 {
		return fmt.Errorf("price per session must be greater than 0")
	}

	if err := s.ValidateSkillLevel(); err != nil {
		return err
	}

	if err := s.ValidateClassType(); err != nil {
		return err
	}

	return nil
}

// ValidateSkillLevel validates skill level
func (s *Service) ValidateSkillLevel() error {
	validLevels := []SkillLevel{
		SkillLevelBeginner,
		SkillLevelIntermediate,
		SkillLevelAdvanced,
		SkillLevelAllLevels,
	}

	for _, level := range validLevels {
		if s.SkillLevel == level {
			return nil
		}
	}

	return fmt.Errorf("invalid skill level: %s", s.SkillLevel)
}

// ValidateClassType validates class type
func (s *Service) ValidateClassType() error {
	validTypes := []ClassType{
		ClassTypePrivate,
		ClassTypeSmallGroup,
		ClassTypeLargeGroup,
	}

	for _, t := range validTypes {
		if s.ClassType == t {
			return nil
		}
	}

	return fmt.Errorf("invalid class type: %s", s.ClassType)
}

// IsActive checks if service is active
func (s *Service) IsActive() bool {
	return s.Status == ServiceStatusActive
}

// IsSuitableForAge checks if service is suitable for a given age
func (s *Service) IsSuitableForAge(age int) bool {
	if s.AgeMin > 0 && age < s.AgeMin {
		return false
	}
	if s.AgeMax > 0 && age > s.AgeMax {
		return false
	}
	return true
}

// GetDefaultPageSize returns default page size for pagination
func GetDefaultPageSize() int {
	return 20
}

// GetMaxPageSize returns maximum page size
func GetMaxPageSize() int {
	return 100
}

// GetAgeRangeDisplay returns formatted age range display
func (s *Service) GetAgeRangeDisplay() string {
	if s.AgeMin > 0 && s.AgeMax > 0 {
		return fmt.Sprintf("%d-%d years", s.AgeMin, s.AgeMax)
	} else if s.AgeMin > 0 {
		return fmt.Sprintf("%d+ years", s.AgeMin)
	} else if s.AgeMax > 0 {
		return fmt.Sprintf("Up to %d years", s.AgeMax)
	}
	return "All ages"
}

// IsValid validates SkillLevel
func (sl SkillLevel) IsValid() bool {
	switch sl {
	case SkillLevelBeginner, SkillLevelIntermediate, SkillLevelAdvanced, SkillLevelAllLevels:
		return true
	}
	return false
}

// IsValid validates ClassType
func (ct ClassType) IsValid() bool {
	switch ct {
	case ClassTypePrivate, ClassTypeSmallGroup, ClassTypeLargeGroup:
		return true
	}
	return false
}

// FromServiceWithAggregates converts ServiceWithAggregates to domain Service
func FromServiceWithAggregates(dm *ServiceWithAggregates) *Service {
	ageMin := 0
	ageMax := 0
	maxParticipants := 0

	if dm.AgeMin != nil {
		ageMin = *dm.AgeMin
	}
	if dm.AgeMax != nil {
		ageMax = *dm.AgeMax
	}
	if dm.MaxParticipants != nil {
		maxParticipants = *dm.MaxParticipants
	}

	domainService := &Service{
		ID:              dm.ID,
		VendorID:        dm.VendorID,
		CategoryID:      dm.CategoryID,
		Name:            dm.Name,
		Description:     dm.Description,
		AgeMin:          ageMin,
		AgeMax:          ageMax,
		SkillLevel:      SkillLevel(dm.SkillLevel),
		ClassType:       ClassType(dm.ClassType),
		MaxParticipants: maxParticipants,
		DurationMinutes: dm.DurationMinutes,
		PricePerSession: dm.PricePerSession,
		IsFeatured:      dm.IsFeatured,
		Status:          ServiceStatus(dm.Status),
		CreatedAt:       dm.CreatedAt,
		UpdatedAt:       dm.UpdatedAt,
		Version:         dm.Version,

		// Aggregated vendor data
		VendorBusinessName: dm.VendorBusinessName,
		VendorCity:         dm.VendorCity,
		VendorDistrict:     dm.VendorDistrict,
		VendorLogo:         dm.VendorLogo,
		VendorRatingAvg:    dm.VendorRatingAvg,
		VendorTotalReviews: dm.VendorTotalReviews,
		VendorVerified:     dm.VendorVerified,
		CategoryName:       dm.CategoryName,
		CategorySlug:       dm.CategorySlug,
	}

	// Handle optional pointer fields - just assign directly
	domainService.TrialPrice = dm.TrialPrice
	domainService.Package4Price = dm.Package4Price
	domainService.Package8Price = dm.Package8Price
	domainService.Package12Price = dm.Package12Price
	domainService.Requirements = dm.Requirements
	domainService.WhatWillLearn = dm.WhatWillLearn

	return domainService
}
