package postgresql

import (
	"context"
)

// GetServiceReviews fetches paginated reviews for a service
func (r *Repository) GetServiceReviews(ctx context.Context, serviceID int64, page, limit int) ([]*ReviewPreviewData, int64, error) {
	offset := (page - 1) * limit

	var reviews []*ReviewPreviewData
	var total int64

	// Count total reviews
	countQuery := `SELECT COUNT(*) FROM reviews WHERE service_id = $1`
	if err := r.db.WithContext(ctx).Raw(countQuery, serviceID).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated reviews (same query as GetTopReviews but with pagination)
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
		LIMIT $2 OFFSET $3
	`
	err := r.db.WithContext(ctx).Raw(query, serviceID, limit, offset).Scan(&reviews).Error

	return reviews, total, err
}
