package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				stack := debug.Stack()
				slog.Error("Panic recovered",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
					slog.Any("panic", rvr),
					slog.String("stack_trace", string(stack)),
				)

				if r.Header.Get("Connection") != "Upgrade" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{
						"error":   "Internal Server Error",
						"message": "An unexpected error occurred",
					})
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
