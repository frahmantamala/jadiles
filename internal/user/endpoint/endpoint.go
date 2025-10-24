package endpoint

import (
	"github.com/frahmantamala/jadiles/internal"
	authpkg "github.com/frahmantamala/jadiles/internal/auth"
	"github.com/frahmantamala/jadiles/internal/user"
	"github.com/frahmantamala/jadiles/internal/user/postgresql"
	"github.com/go-chi/chi/v5"
	goRedis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(
	r chi.Router,
	db *gorm.DB,
	redisClient goRedis.UniversalClient,
	config internal.Config,
) error {
	jwtAuth, err := authpkg.NewJWTAuthentication(config.HTTPServer)
	if err != nil {
		return err
	}

	passwordManager := authpkg.NewPasswordManager()

	tokenStorage := authpkg.NewRedisTokenStorage(redisClient)

	repo := postgresql.NewUserRepository(db)

	userService := user.NewService(repo, jwtAuth, passwordManager, tokenStorage)

	userHandler := user.NewHandler(userService)
	// Public routes (no authentication required)
	r.Group(func(r chi.Router) {
		r.Post("/register/parent", userHandler.RegisterParent)
		r.Post("/register/vendor", userHandler.RegisterVendor)
		r.Post("/login", userHandler.Login)
		r.Post("/refresh", userHandler.RefreshToken)
	})

	r.Group(func(r chi.Router) {
		r.Use(jwtAuth.Authenticator)

		r.Post("/logout", userHandler.Logout)
		r.Get("/me", userHandler.GetProfile)
	})

	return nil
}
