package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = Default()
	SetDefault(defaultLogger)
}

type Config struct {
	Level  slog.Level
	Writer io.Writer
}

func New(cfg Config) *Logger {
	if cfg.Writer == nil {
		cfg.Writer = os.Stdout
	}

	handler := slog.NewJSONHandler(cfg.Writer, &slog.HandlerOptions{
		Level: cfg.Level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			if a.Key == slog.MessageKey {
				a.Key = "message"
			}
			if a.Key == slog.LevelKey {
				a.Key = "level"
			}
			return a
		},
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

func Default() *Logger {
	return New(Config{
		Level:  slog.LevelInfo,
		Writer: os.Stdout,
	})
}

func SetDefault(l *Logger) {
	slog.SetDefault(l.Logger)
}

func Alert(err error) slog.Attr {
	return slog.Group("error",
		slog.Bool("alert", true),
		slog.String("message", err.Error()),
	)
}

func ErrorDetails(err error, details map[string]any) slog.Attr {
	attrs := []any{
		slog.String("message", err.Error()),
	}
	for k, v := range details {
		attrs = append(attrs, slog.Any(k, v))
	}
	return slog.Group("error", attrs...)
}

func WithContext(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}

	if reqID := ctx.Value("request_id"); reqID != nil {
		attrs = append(attrs, slog.String("request_id", reqID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		attrs = append(attrs, slog.Any("user_id", userID))
	}

	return attrs
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	slog.DebugContext(ctx, msg, args...)
}

var (
	sensitiveDataField = []string{
		"password",
		"client_secret",
		"token",
		"access_token",
		"refresh_token",
		"accessToken",
		"refreshToken",
		"zendesk_jwt_token",
		"account_number",
		"accountNumber",
		"account_name",
		"account_holder_name",
		"accountHolderName",
		"company_sso_id",
		"owner_sso_id",
		"userSSOID",
		"companyName",
		"ownerName",
		"ownerEmail",
		"ownerPhone",
		"ownerPassword",
		"ownerConfirmPassword",
		"phone",
		"address",
		"email",
	}

	filterDataField = []string{"cardCVV", "cardNumber", "cardExpiryDate"}

	SensitiveValueMatcher = func() KeyMatcher {
		matcher := make(map[string]ValueObfuscator)
		for _, key := range sensitiveDataField {
			matcher[key] = ValueSHA256Hashed()
		}
		for _, key := range filterDataField {
			matcher[key] = ValueToFixedValue("[MASKED]")
		}
		return KeyMatcherExact(matcher)
	}()

	LogAttributeFmter = HideSensitiveValueOnKeys(SensitiveValueMatcher)
)
