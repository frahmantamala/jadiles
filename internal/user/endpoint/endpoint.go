package endpoint

// import (
// 	"github.com/frahmantamala/jadiles/internal/user"
// 	"github.com/frahmantamala/jadiles/internal/user/postgresql"
// 	"github.com/go-chi/chi/v5"
// 	"gorm.io/gorm"
// )

// func NewWebEndpoint(
// 	db *gorm.DB,
// ) *chi.Mux {
// 	mux := chi.NewMux()
// 	repo := postgresql.NewAuthRepository(db)

// 	userSvc := user.NewService(repo)

// 	userHandler := user.NewHandler(userSvc)

// 	mux.Post("/logout", userHandler.Logout)

// 	return mux
// }
