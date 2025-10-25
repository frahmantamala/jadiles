package endpoint

import (
	"github.com/frahmantamala/jadiles/internal/services"
	"github.com/frahmantamala/jadiles/internal/services/postgresql"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// RegisterServiceRoutes registers service-related routes
func RegisterServiceRoutes(r chi.Router, db *gorm.DB) error {
	// Initialize repository
	repo := postgresql.NewRepository(db)

	// Initialize service (usecase)
	svc := services.NewService(repo)

	// Initialize handler
	handler := services.NewHandler(svc)

	// Public routes (no authentication required)
	r.Get("/services/search", handler.SearchServices)
	r.Get("/categories", handler.GetCategories)

	return nil
}
