package transport

import (
	"embed"
	"net/http"

	"github.com/frahmantamala/jadiles/internal"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/chi/v5"
)

//go:embed swagger
var swaggerFiles embed.FS

func swaggerRoutes(r *chi.Mux, cfg internal.SwaggerConfig) {
	if !cfg.Enable {
		return
	}

	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.FS(swaggerFiles)),
	))

	r.Get("/openapi3.json", func(w http.ResponseWriter, r *http.Request) {
		s, err := v1.GetSwagger()
		if err != nil {
			return
		}
		b, _ := s.MarshalJSON()

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	})
}
