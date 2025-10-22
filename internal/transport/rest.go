package transport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authEndpoint "github.com/frahmantamala/jadiles/internal/auth/endpoint"
	"github.com/go-chi/chi/v5"

	"github.com/frahmantamala/jadiles/internal"
	"github.com/frahmantamala/jadiles/pkg/logger"
	"github.com/gomodule/redigo/redis"
	goRedis "github.com/redis/go-redis/v9"
	chitrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-chi/chi.v5"
	"gorm.io/gorm"
)

const (
	defaultHealthCheckTimeout = 2 * time.Second
)

type RESTServer struct {
	srv *http.Server
}

func NewRESTServer(
	gormDB *gorm.DB,
	goRedisClient goRedis.UniversalClient,
	redisConn *redis.Pool,
	config internal.Config,
) (*RESTServer, error) {

	// Initialize logger middleware
	logMw, err := logger.LogMiddleware(
		logger.LoggerMwOption().
			WithLogger(slog.Default()).
			WithRequestBodyDecoder(logger.ByteDecoderJSONObfuscator(logger.SensitiveValueMatcher)).
			WithResponseBodyDecoder(logger.ByteDecoderJSONObfuscator(logger.SensitiveValueMatcher)).
			WithAllowedHTTPStatusesResponse(logger.HTTPStatus5xx, logger.HTTPStatus4xx).
			WithSkipPath("/v1/media/upload"),
	)
	if err != nil {
		return nil, err
	}

	routes := chi.NewRouter()
	routes.Use(CORSMiddleware(config.HTTPServer.GetAllowedOrigins()))
	routes.Use(Recoverer)
	routes.Use(chitrace.Middleware(chitrace.WithServiceName(config.Name)), TraceIDHandler)

	routes.Get("/ping", pingHandler)
	routes.Get(
		"/health",
		healthCheckHandler(defaultHealthCheckTimeout, gormDB, redisConn),
	)

	var routeErr error
	routes.Route("/v1", func(v1 chi.Router) {
		v1.Use(StripSlashes)

		// Apply logger middleware for all v1 routes
		v1.Group(func(r chi.Router) {
			r.Use(logMw.Middleware)

			// Register auth routes
			if err := authEndpoint.RegisterAuthRoutes(r, gormDB, goRedisClient, config); err != nil {
				routeErr = fmt.Errorf("failed to register auth routes: %w", err)
				return
			}
		})
	})

	if routeErr != nil {
		return nil, routeErr
	}

	swaggerRoutes(routes, config.Swagger)

	return &RESTServer{
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", config.HTTPServer.Port),
			Handler:           routes,
			ReadTimeout:       config.HTTPServer.ReadTimeout,
			ReadHeaderTimeout: config.HTTPServer.ReadHeaderTimeout,
			IdleTimeout:       config.HTTPServer.IdleTimeout,
			WriteTimeout:      config.HTTPServer.WriteTimeout,
		},
	}, nil
}

func (r *RESTServer) Start() error {
	return r.srv.ListenAndServe()
}

func (r *RESTServer) Stop(ctx context.Context) error {
	return r.srv.Shutdown(ctx)
}
