package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/frahmantamala/jadiles/internal/core/datamodel"
	"github.com/frahmantamala/jadiles/internal/services"
)

// ServiceDetailData holds all data for service detail view
type ServiceDetailData struct {
	Service        *datamodel.Services
	Vendor         *VendorDetailData
	Coaches        []*CoachDetailData
	Schedules      []*ScheduleDetailData
	ReviewsPreview []*ReviewPreviewData
	ReviewSummary  *ReviewSummaryData
}

// VendorDetailData represents vendor with photos and amenities
type VendorDetailData struct {
	ID           int64
	BusinessName string
	Description  *string
	Phone        string
	WhatsApp     *string
	Address      string
	City         string
	District     string
	Photos       *string  // JSONB string
	Amenities    *string  // JSONB string
	Logo         *string
	CoverImage   *string
	RatingAvg    *float64
	TotalReviews int
	Verified     bool
}

// CoachDetailData represents coach with certifications
type CoachDetailData struct {
	ID              int64
	FullName        string
	Bio             *string
	Photo           *string
	ExperienceYears int
	Education       *string
	Certifications  *string // JSONB string
	Specializations *string // JSONB string
	IsPrimary       bool
	IsFeatured      bool
}

// ScheduleDetailData represents schedule with coach name
type ScheduleDetailData struct {
	ID             int64
	DayOfWeek      int
	StartTime      string
	EndTime        string
	AvailableSlots int
	CoachID        *int64
	CoachName      *string
	IsActive       bool
}

// ReviewPreviewData represents review with parent info
type ReviewPreviewData struct {
	ID             int64
	ParentName     string
	ChildAge       *int
	Rating         int
	ReviewText     *string
	DidChildEnjoy  *bool
	WouldRecommend bool
	Photos         *string // JSONB string
	VendorResponse *string
	RespondedAt    *sql.NullTime
	CreatedAt      sql.NullTime
}

// ReviewSummaryData represents aggregated review statistics
type ReviewSummaryData struct {
	TotalReviews        int
	AverageRating       float64
	Rating1             int
	Rating2             int
	Rating3             int
	Rating4             int
	Rating5             int
	ChildEnjoyedPct     float64
	WouldRecommendPct   float64
}

// GetServiceDetail fetches all service detail data sequentially
func (r *Repository) GetServiceDetail(ctx context.Context, serviceID int64) (*ServiceDetailData, error) {
	// 1. Get service
	service, err := r.GetServiceByID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// 2. Get vendor with details
	vendor, err := r.GetVendorByServiceID(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// 3. Get coaches assigned to this service
	coaches, err := r.GetServiceCoaches(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// 4. Get schedules
	schedules, err := r.GetServiceSchedules(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	// 5. Get top 3 recent reviews
	reviewsPreview, err := r.GetTopReviews(ctx, serviceID, 3)
	if err != nil {
		return nil, err
	}

	// 6. Get review statistics
	reviewSummary, err := r.GetReviewSummary(ctx, serviceID)
	if err != nil {
		return nil, err
	}

	return &ServiceDetailData{
		Service:        service,
		Vendor:         vendor,
		Coaches:        coaches,
		Schedules:      schedules,
		ReviewsPreview: reviewsPreview,
		ReviewSummary:  reviewSummary,
	}, nil
}

// GetVendorByServiceID fetches vendor details with photos and amenities
func (r *Repository) GetVendorByServiceID(ctx context.Context, serviceID int64) (*VendorDetailData, error) {
	var vendor VendorDetailData
	query := `
		SELECT
			v.id, v.business_name, v.description, v.phone, v.whatsapp,
			v.address, v.city, v.district, v.photos, v.amenities,
			v.logo, v.cover_image, v.rating_avg, v.total_reviews, v.verified
		FROM vendors v
		INNER JOIN services s ON v.id = s.vendor_id
		WHERE s.id = $1 AND v.status = 'approved'
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&vendor).Error
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

// GetServiceCoaches fetches coaches with join table information
func (r *Repository) GetServiceCoaches(ctx context.Context, serviceID int64) ([]*CoachDetailData, error) {
	var coaches []*CoachDetailData
	query := `
		SELECT
			c.id, c.full_name, c.bio, c.photo, c.experience_years,
			c.education, c.certifications, c.specializations, c.is_featured,
			sc.is_primary
		FROM coaches c
		INNER JOIN service_coaches sc ON c.id = sc.coach_id
		WHERE sc.service_id = $1 AND c.status = 'active'
		ORDER BY sc.is_primary DESC, c.is_featured DESC, c.full_name ASC
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&coaches).Error
	return coaches, err
}

// GetServiceSchedules fetches schedules with coach names
func (r *Repository) GetServiceSchedules(ctx context.Context, serviceID int64) ([]*ScheduleDetailData, error) {
	var schedules []*ScheduleDetailData
	query := `
		SELECT
			s.id, s.day_of_week, s.start_time, s.end_time, s.available_slots,
			s.coach_id, c.full_name as coach_name, s.is_active
		FROM schedules s
		LEFT JOIN coaches c ON s.coach_id = c.id
		WHERE s.service_id = $1 AND s.is_active = true
		ORDER BY s.day_of_week ASC, s.start_time ASC
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&schedules).Error
	return schedules, err
}

// GetTopReviews fetches recent reviews with parent names and child age
func (r *Repository) GetTopReviews(ctx context.Context, serviceID int64, limit int) ([]*ReviewPreviewData, error) {
	var reviews []*ReviewPreviewData
	query := `
		SELECT
			r.id, u.full_name as parent_name, r.rating, r.review_text,
			r.did_child_enjoy, r.would_recommend, r.photos,
			r.vendor_response, r.vendor_responded_at, r.created_at,
			EXTRACT(YEAR FROM AGE(CURRENT_DATE, ch.date_of_birth))::int as child_age
		FROM reviews r
		INNER JOIN bookings b ON r.booking_id = b.id
		INNER JOIN users u ON r.parent_id = u.id
		INNER JOIN children ch ON b.child_id = ch.id
		WHERE r.service_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID, limit).Scan(&reviews).Error
	return reviews, err
}

// GetReviewSummary calculates aggregate review statistics
func (r *Repository) GetReviewSummary(ctx context.Context, serviceID int64) (*ReviewSummaryData, error) {
	var summary ReviewSummaryData
	query := `
		SELECT
			COUNT(*) as total_reviews,
			COALESCE(AVG(rating), 0) as average_rating,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) as rating_1,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) as rating_2,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) as rating_3,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) as rating_4,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) as rating_5,
			COALESCE(AVG(CASE WHEN did_child_enjoy = true THEN 100.0 ELSE 0.0 END), 0) as child_enjoyed_pct,
			COALESCE(AVG(CASE WHEN would_recommend = true THEN 100.0 ELSE 0.0 END), 0) as would_recommend_pct
		FROM reviews
		WHERE service_id = $1
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID).Scan(&summary).Error
	return &summary, err
}

// ToServiceDetail converts repository data to domain ServiceDetail
func ToServiceDetail(data *ServiceDetailData) (*services.ServiceDetail, error) {
	// Convert basic service
	service := services.FromServiceWithAggregates(&services.ServiceWithAggregates{
		Services: *data.Service,
	})

	// Convert vendor
	vendor, err := toVendorDetail(data.Vendor)
	if err != nil {
		return nil, err
	}

	// Convert coaches
	coaches, err := toCoachDetails(data.Coaches)
	if err != nil {
		return nil, err
	}

	// Convert schedules
	schedules := toScheduleDetails(data.Schedules)

	// Convert reviews preview
	reviewsPreview, err := toReviewPreviews(data.ReviewsPreview)
	if err != nil {
		return nil, err
	}

	// Convert review summary
	reviewSummary := toReviewSummary(data.ReviewSummary)

	return &services.ServiceDetail{
		Service:        *service,
		Vendor:         vendor,
		Coaches:        coaches,
		Schedules:      schedules,
		ReviewsPreview: reviewsPreview,
		ReviewSummary:  reviewSummary,
	}, nil
}

func toVendorDetail(data *VendorDetailData) (*services.VendorDetail, error) {
	var photos []string
	var amenities []string

	// Parse JSONB photos
	if data.Photos != nil {
		if err := json.Unmarshal([]byte(*data.Photos), &photos); err != nil {
			photos = []string{}
		}
	}

	// Parse JSONB amenities
	if data.Amenities != nil {
		if err := json.Unmarshal([]byte(*data.Amenities), &amenities); err != nil {
			amenities = []string{}
		}
	}

	return &services.VendorDetail{
		ID:           data.ID,
		BusinessName: data.BusinessName,
		Description:  data.Description,
		Phone:        data.Phone,
		WhatsApp:     data.WhatsApp,
		Address:      data.Address,
		City:         data.City,
		District:     data.District,
		Photos:       photos,
		Amenities:    amenities,
		Logo:         data.Logo,
		CoverImage:   data.CoverImage,
		RatingAvg:    data.RatingAvg,
		TotalReviews: data.TotalReviews,
		Verified:     data.Verified,
	}, nil
}

func toCoachDetails(data []*CoachDetailData) ([]*services.CoachDetail, error) {
	coaches := make([]*services.CoachDetail, 0, len(data))
	for _, c := range data {
		var certifications []services.Certification
		var specializations []string

		// Parse JSONB certifications
		if c.Certifications != nil {
			if err := json.Unmarshal([]byte(*c.Certifications), &certifications); err != nil {
				certifications = []services.Certification{}
			}
		}

		// Parse JSONB specializations
		if c.Specializations != nil {
			if err := json.Unmarshal([]byte(*c.Specializations), &specializations); err != nil {
				specializations = []string{}
			}
		}

		coaches = append(coaches, &services.CoachDetail{
			ID:              c.ID,
			FullName:        c.FullName,
			Bio:             c.Bio,
			Photo:           c.Photo,
			ExperienceYears: c.ExperienceYears,
			Education:       c.Education,
			Certifications:  certifications,
			Specializations: specializations,
			IsPrimary:       c.IsPrimary,
			IsFeatured:      c.IsFeatured,
		})
	}
	return coaches, nil
}

func toScheduleDetails(data []*ScheduleDetailData) []*services.ScheduleDetail {
	schedules := make([]*services.ScheduleDetail, 0, len(data))
	for _, s := range data {
		schedules = append(schedules, &services.ScheduleDetail{
			ID:             s.ID,
			DayOfWeek:      s.DayOfWeek,
			DayName:        services.GetDayName(s.DayOfWeek),
			StartTime:      s.StartTime,
			EndTime:        s.EndTime,
			AvailableSlots: s.AvailableSlots,
			CoachID:        s.CoachID,
			CoachName:      s.CoachName,
			IsActive:       s.IsActive,
		})
	}
	return schedules
}

func toReviewPreviews(data []*ReviewPreviewData) ([]*services.ReviewPreview, error) {
	reviews := make([]*services.ReviewPreview, 0, len(data))
	for _, r := range data {
		var photos []string

		// Parse JSONB photos
		if r.Photos != nil {
			if err := json.Unmarshal([]byte(*r.Photos), &photos); err != nil {
				photos = []string{}
			}
		}

		var respondedAt *time.Time
		if r.RespondedAt != nil && r.RespondedAt.Valid {
			t := r.RespondedAt.Time
			respondedAt = &t
		}

		var createdAt time.Time
		if r.CreatedAt.Valid {
			createdAt = r.CreatedAt.Time
		}

		reviews = append(reviews, &services.ReviewPreview{
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
		})
	}
	return reviews, nil
}

func toReviewSummary(data *ReviewSummaryData) *services.ReviewSummary {
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
