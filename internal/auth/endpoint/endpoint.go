package endpoint

import (
	"fmt"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/internal/auth"
	"github.com/frahmantamala/jadiles/internal/auth/postgresql"
	authRedis "github.com/frahmantamala/jadiles/internal/auth/redis"
	"github.com/go-chi/chi/v5"
	goRedis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterAuthRoutes registers all authentication routes
// Dependencies are injected from the HTTP server
func RegisterAuthRoutes(r chi.Router, db *gorm.DB, redisClient goRedis.UniversalClient, config internal.Config) error {
	// Get JWT secret from config
	jwtSecret, err := config.HTTPServer.GetJWTSecret()
	if err != nil {
		return fmt.Errorf("failed to decode JWT secret: %w", err)
	}

	// Initialize dependencies
	repo := postgresql.NewAuthRepository(db)
	jwtManager := auth.NewJWTManager(
		string(jwtSecret),
		config.HTTPServer.AuthConfig.AccessTokenDuration,
		config.HTTPServer.AuthConfig.RefreshTokenDuration,
	)
	passwordManager := auth.NewPasswordManager()
	tokenStorage := authRedis.NewTokenStorage(redisClient)

	// Initialize service
	service := auth.NewService(repo, jwtManager, passwordManager, tokenStorage)

	// Initialize handler
	handler := auth.NewHandler(service)

	// Register routes
	r.Post("/auth/register/parent", handler.RegisterParent)
	r.Post("/auth/register/vendor", handler.RegisterVendor)
	r.Post("/auth/login", handler.Login)

	// Protected routes (will need auth middleware)
	r.Post("/auth/logout", handler.Logout) // TODO: Add auth middleware

	return nil
}
