package internal

import (
	"context"
	"log/slog"

	"github.com/frahmantamala/jadiles/pkg/logger"
)

const (
	AppCtxKey = "app_context"
)

type SlogCtx struct {
	DDTraceID string
	DDSpanID  string
	UserID    int64
}

func (s *SlogCtx) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("dd_trace_id", s.DDTraceID),
		slog.String("dd_span_id", s.DDSpanID),
		slog.Int64("user_id", s.UserID),
	)
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

	SensitiveValueMatcher = func() logger.KeyMatcher {
		matcher := make(map[string]logger.ValueObfuscator)
		for _, key := range sensitiveDataField {
			matcher[key] = logger.ValueSHA256Hashed()
		}
		for _, key := range filterDataField {
			matcher[key] = logger.ValueToFixedValue("[MASKED]")
		}
		return logger.KeyMatcherExact(matcher)
	}()

	// For HTTP middleware (JSON body obfuscation)
	LogBodyObfuscator = logger.HideSensitiveValueOnKeys(SensitiveValueMatcher)
)

// LogAttributeFmter is a function (not a variable) for slog attribute formatting
func LogAttributeFmter(groups []string, attr slog.Attr) slog.Attr {
	// Check if this attribute key matches any sensitive field
	if obfuscator, matches := SensitiveValueMatcher.Match(attr.Key); matches {
		// Get the value as interface{}
		val := attr.Value.Any()
		// Apply obfuscation
		obfuscated := obfuscator(val)
		// Return new attribute with obfuscated value
		return slog.Any(attr.Key, obfuscated)
	}

	// If it's a group, recursively process its attributes
	if attr.Value.Kind() == slog.KindGroup {
		groupAttrs := attr.Value.Group()
		newAttrs := make([]slog.Attr, 0, len(groupAttrs))

		// Add current key to groups path
		newGroups := append(groups, attr.Key)

		for _, a := range groupAttrs {
			newAttrs = append(newAttrs, LogAttributeFmter(newGroups, a))
		}

		return slog.Attr{
			Key:   attr.Key,
			Value: slog.GroupValue(newAttrs...),
		}
	}

	return attr
}

func SlogContextExtractor(ctx context.Context) []slog.Attr {
	attrs := []slog.Attr{}

	if reqID := ctx.Value("request_id"); reqID != nil {
		attrs = append(attrs, slog.String("request_id", reqID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		attrs = append(attrs, slog.Any("user_id", userID))
	}

	return attrs
}
