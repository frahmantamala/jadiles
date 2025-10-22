package transport

import (
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	t "github.com/frahmantamala/jadiles/internal/tracer"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	defautCORSMaxAge = 300
)

func CORSMiddleware(origins []string) func(handler http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Authenticated-Userid",
			"x-datadog-trace-id",
			"x-datadog-parent-id",
			"x-datadog-origin",
			"x-datadog-sampling-priority",
		},
		ExposedHeaders:     []string{"Link"},
		AllowCredentials:   false,
		MaxAge:             defautCORSMaxAge,
		OptionsPassthrough: false,
		Debug:              false,
	})
}

func TraceIDHandler(next http.Handler) http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.NewString()
		if span, ok := tracer.SpanFromContext(r.Context()); ok {
			span.SetTag(t.ApplicationTraceIDKey, traceID)
		}
		ctx := internal.InjectTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}))
}
