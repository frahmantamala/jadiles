package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/gomodule/redigo/redis"
	"github.com/hellofresh/health-go/v5"
	"gorm.io/gorm"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, struct {
		Status string `json:"status"`
	}{
		Status: "OK",
	})
}

func healthCheckHandler(
	timeout time.Duration,
	gormDB *gorm.DB,
	redisPool *redis.Pool,
) http.HandlerFunc {
	h, _ := health.New(
		health.WithComponent(health.Component{
			Name:    "jadiles",
			Version: "",
		}),
		health.WithMaxConcurrent(1),
	)

	_ = h.Register(health.Config{
		Name:      "postgres",
		Timeout:   timeout,
		SkipOnErr: false,
		Check: func(ctx context.Context) error {
			db, err := gormDB.DB()
			if err != nil {
				return err
			}
			return db.PingContext(ctx)
		},
	})

	_ = h.Register(health.Config{
		Name:      "redis",
		Timeout:   timeout,
		SkipOnErr: false,
		Check: func(ctx context.Context) error {
			conn := redisPool.Get()
			defer conn.Close()

			_, err := conn.Do("PING")
			return err
		},
	})

	return h.HandlerFunc
}
