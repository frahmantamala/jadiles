package endpoint

import (
	authpkg "github.com/frahmantamala/jadiles/internal/auth"
	"github.com/frahmantamala/jadiles/internal/child"
	"github.com/frahmantamala/jadiles/internal/child/postgresql"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// RegisterChildRoutes registers all child-related routes
func RegisterChildRoutes(
	r chi.Router,
	db *gorm.DB,
	jwtAuth *authpkg.JWTAuthentication,
) error {
	repo := postgresql.NewChildRepository(db)
	childService := child.NewService(repo)
	childHandler := child.NewHandler(childService)

	// Protected routes (require authentication and parent role)
	r.Group(func(r chi.Router) {
		r.Use(jwtAuth.Authenticator)
		r.Use(jwtAuth.RequireRole("parent"))

		r.Get("/children", childHandler.GetChildren)
		r.Post("/children", childHandler.AddChild)
		r.Get("/children/{id}", childHandler.GetChild)
		r.Put("/children/{id}", childHandler.UpdateChild)
		r.Delete("/children/{id}", childHandler.DeleteChild)
	})

	return nil
}
