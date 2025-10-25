package endpoint

import (
	"github.com/frahmantamala/jadiles/internal/services/booking"
	"github.com/frahmantamala/jadiles/internal/services/detail"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	"github.com/frahmantamala/jadiles/internal/services/review"
	"github.com/frahmantamala/jadiles/internal/services/schedule"
	"github.com/frahmantamala/jadiles/internal/services/search"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// RegisterServiceRoutes registers service-related routes
func RegisterServiceRoutes(r chi.Router, db *gorm.DB) error {
	// Initialize repository
	repo := postgresql.NewRepository(db)

	// Initialize search capability
	searchSvc := search.NewService(repo)
	searchHandler := search.NewHandler(searchSvc)

	// Initialize detail capability
	detailSvc := detail.NewService(repo)
	detailHandler := detail.NewHandler(detailSvc)

	// Initialize schedule capability
	scheduleSvc := schedule.NewService(repo)
	scheduleHandler := schedule.NewHandler(scheduleSvc)

	// Initialize review capability
	reviewSvc := review.NewService(repo)
	reviewHandler := review.NewHandler(reviewSvc)

	// Initialize booking capability
	bookingSvc := booking.NewService(repo)
	bookingHandler := booking.NewHandler(bookingSvc)

	// Public routes (no authentication required)
	r.Get("/services/search", searchHandler.SearchServices)
	r.Get("/categories", searchHandler.GetCategories)
	r.Get("/services/{service_id}", detailHandler.GetServiceDetail)
	r.Get("/services/{service_id}/availability", scheduleHandler.GetServiceAvailability)
	r.Get("/services/{service_id}/reviews", reviewHandler.GetServiceReviews)

	// Booking routes (require authentication - TODO: add auth middleware)
	r.Post("/bookings", bookingHandler.CreateBooking)
	r.Get("/bookings/{booking_id}", bookingHandler.GetBooking)

	return nil
}
