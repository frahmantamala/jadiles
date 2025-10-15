package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// SlogOption holds configuration for creating a custom slog handler
type SlogOption struct {
	Resource           map[string]string
	ContextExtractor   func(context.Context) []slog.Attr
	AttributeFormatter func([]string, slog.Attr) slog.Attr
	Writer             io.Writer
	Leveler            slog.Leveler
}

// NewHandler creates a new slog.Handler with the configured options
func (opt SlogOption) NewHandler() slog.Handler {
	if opt.Writer == nil {
		opt.Writer = os.Stdout
	}

	if opt.Leveler == nil {
		opt.Leveler = slog.LevelInfo
	}

	// Create base handler with standard options
	handlerOpts := &slog.HandlerOptions{
		Level: opt.Leveler,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Standard attribute name replacements
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}

			// Apply custom attribute formatter if provided
			if opt.AttributeFormatter != nil {
				a = opt.AttributeFormatter(groups, a)
			}

			return a
		},
	}

	baseHandler := slog.NewJSONHandler(opt.Writer, handlerOpts)

	// Wrap with custom handler that adds resource attributes and context extraction
	return &customHandler{
		Handler:          baseHandler,
		resource:         opt.Resource,
		contextExtractor: opt.ContextExtractor,
	}
}

// customHandler wraps slog.Handler to add resource attributes and context extraction
type customHandler struct {
	slog.Handler
	resource         map[string]string
	contextExtractor func(context.Context) []slog.Attr
}

// Handle processes a log record, adding resource attributes and context data
func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	// Add resource attributes
	for key, value := range h.resource {
		r.AddAttrs(slog.String(key, value))
	}

	// Extract and add context attributes if extractor is provided
	if h.contextExtractor != nil {
		contextAttrs := h.contextExtractor(ctx)
		for _, attr := range contextAttrs {
			r.AddAttrs(attr)
		}
	}

	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new handler with the given attributes added
func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customHandler{
		Handler:          h.Handler.WithAttrs(attrs),
		resource:         h.resource,
		contextExtractor: h.contextExtractor,
	}
}

// WithGroup returns a new handler with the given group added
func (h *customHandler) WithGroup(name string) slog.Handler {
	return &customHandler{
		Handler:          h.Handler.WithGroup(name),
		resource:         h.resource,
		contextExtractor: h.contextExtractor,
	}
}
