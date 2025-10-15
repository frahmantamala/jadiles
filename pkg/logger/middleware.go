package logger

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HTTPStatusCode represents HTTP status code ranges
type HTTPStatusCode string

const (
	HTTPStatus1xx HTTPStatusCode = "1xx"
	HTTPStatus2xx HTTPStatusCode = "2xx"
	HTTPStatus3xx HTTPStatusCode = "3xx"
	HTTPStatus4xx HTTPStatusCode = "4xx"
	HTTPStatus5xx HTTPStatusCode = "5xx"
)

// ByteDecoder is a function type that decodes and potentially obfuscates byte data
type ByteDecoder func([]byte) string

// LogMwOption holds configuration for the logging middleware
type LogMwOption struct {
	logger                       *slog.Logger
	httpReqBodyDecoder           ByteDecoder
	httpRespBodyDecoder          ByteDecoder
	allowedHTTPStatusesResponse  map[HTTPStatusCode]bool
	excludedHTTPStatusesResponse map[HTTPStatusCode]bool
	skipPath                     map[string]bool
}

// LoggerMwOptionBuilder provides a builder pattern for middleware configuration
type LoggerMwOptionBuilder struct {
	option *LogMwOption
}

// LoggerMwOption creates a new option builder
func LoggerMwOption() *LoggerMwOptionBuilder {
	return &LoggerMwOptionBuilder{
		option: &LogMwOption{
			logger:                       slog.Default(),
			allowedHTTPStatusesResponse:  make(map[HTTPStatusCode]bool),
			excludedHTTPStatusesResponse: make(map[HTTPStatusCode]bool),
			skipPath:                     make(map[string]bool),
		},
	}
}

// WithLogger sets the logger instance
func (b *LoggerMwOptionBuilder) WithLogger(logger *slog.Logger) *LoggerMwOptionBuilder {
	b.option.logger = logger
	return b
}

// WithRequestBodyDecoder sets the request body decoder
func (b *LoggerMwOptionBuilder) WithRequestBodyDecoder(decoder ByteDecoder) *LoggerMwOptionBuilder {
	b.option.httpReqBodyDecoder = decoder
	return b
}

// WithResponseBodyDecoder sets the response body decoder
func (b *LoggerMwOptionBuilder) WithResponseBodyDecoder(decoder ByteDecoder) *LoggerMwOptionBuilder {
	b.option.httpRespBodyDecoder = decoder
	return b
}

// WithAllowedHTTPStatusesResponse sets which HTTP status ranges should log response bodies
func (b *LoggerMwOptionBuilder) WithAllowedHTTPStatusesResponse(statuses ...HTTPStatusCode) *LoggerMwOptionBuilder {
	for _, status := range statuses {
		b.option.allowedHTTPStatusesResponse[status] = true
	}
	return b
}

// WithExcludedHTTPStatusesResponse sets which HTTP status ranges should NOT log response bodies
func (b *LoggerMwOptionBuilder) WithExcludedHTTPStatusesResponse(statuses ...HTTPStatusCode) *LoggerMwOptionBuilder {
	for _, status := range statuses {
		b.option.excludedHTTPStatusesResponse[status] = true
	}
	return b
}

// WithSkipPath adds a path to skip logging
func (b *LoggerMwOptionBuilder) WithSkipPath(path string) *LoggerMwOptionBuilder {
	b.option.skipPath[path] = true
	return b
}

// LogMw is the logging middleware
type LogMw struct {
	cfg *LogMwOption
}

// LogMiddleware creates a new logging middleware with the given configuration
func LogMiddleware(builder *LoggerMwOptionBuilder) (*LogMw, error) {
	if builder == nil {
		builder = LoggerMwOption()
	}

	return &LogMw{
		cfg: builder.option,
	}, nil
}

// Middleware returns the HTTP middleware handler
func (mw *LogMw) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip path that doesn't need to be logged
		if mw.cfg.skipPath != nil {
			if _, exist := mw.cfg.skipPath[r.URL.Path]; exist {
				next.ServeHTTP(w, r)
				return
			}
		}

		t0 := time.Now()

		// Capture request
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Log request
		mw.logRequest(r, requestBody)

		// Create response recorder
		respRec := &customResponseWriterImpl{
			ResponseWriter: w,
		}

		// Call next handler
		next.ServeHTTP(respRec, r)

		// Calculate elapsed time
		elapsed := time.Since(t0)

		// Log response
		mw.logResponse(r, respRec, elapsed)
	})
}

// logRequest logs the incoming request
func (mw *LogMw) logRequest(r *http.Request, body []byte) {
	attrs := []any{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("type", "request"),
	}

	// Add request body if decoder is provided
	if mw.cfg.httpReqBodyDecoder != nil && len(body) > 0 {
		attrs = append(attrs, slog.String("body", mw.cfg.httpReqBodyDecoder(body)))
	}

	mw.cfg.logger.Info("HTTP request", attrs...)
}

// logResponse logs the outgoing response
func (mw *LogMw) logResponse(r *http.Request, respRec *customResponseWriterImpl, elapsed time.Duration) {
	statusCode := respRec.statusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	// Determine if we should log response body based on status code
	shouldLogResponse := mw.shouldLogResponseBody(statusCode)

	attrs := []any{
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("remote_addr", r.RemoteAddr),
		slog.Int("status", statusCode),
		slog.Duration("elapsed_time", elapsed),
		slog.String("type", "response"),
	}

	// Add response body if decoder is provided and status matches
	if mw.cfg.httpRespBodyDecoder != nil && shouldLogResponse && len(respRec.body) > 0 {
		attrs = append(attrs, slog.String("body", mw.cfg.httpRespBodyDecoder(respRec.body)))
	}

	// Log with appropriate level based on status code
	if statusCode >= 500 {
		mw.cfg.logger.Error("HTTP response", attrs...)
	} else if statusCode >= 400 {
		mw.cfg.logger.Warn("HTTP response", attrs...)
	} else {
		mw.cfg.logger.Info("HTTP response", attrs...)
	}
}

// shouldLogResponseBody determines if response body should be logged based on status code
func (mw *LogMw) shouldLogResponseBody(statusCode int) bool {
	// Get status range
	var statusRange HTTPStatusCode
	switch {
	case statusCode >= 100 && statusCode < 200:
		statusRange = HTTPStatus1xx
	case statusCode >= 200 && statusCode < 300:
		statusRange = HTTPStatus2xx
	case statusCode >= 300 && statusCode < 400:
		statusRange = HTTPStatus3xx
	case statusCode >= 400 && statusCode < 500:
		statusRange = HTTPStatus4xx
	case statusCode >= 500 && statusCode < 600:
		statusRange = HTTPStatus5xx
	}

	// Check if explicitly excluded
	if mw.cfg.excludedHTTPStatusesResponse[statusRange] {
		return false
	}

	// Check if explicitly allowed
	if len(mw.cfg.allowedHTTPStatusesResponse) > 0 {
		return mw.cfg.allowedHTTPStatusesResponse[statusRange]
	}

	// Default: don't log response body
	return false
}

// customResponseWriterImpl wraps http.ResponseWriter to capture response data
type customResponseWriterImpl struct {
	http.ResponseWriter
	statusCode int
	body       []byte
	header     http.Header
}

func (w *customResponseWriterImpl) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *customResponseWriterImpl) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

func (w *customResponseWriterImpl) Header() http.Header {
	if w.header == nil {
		w.header = w.ResponseWriter.Header()
	}
	return w.header
}
