package internal

import (
	"context"
	"errors"
	"time"
)

type contextKey string

const (
	UserIDCtxKey    = contextKey("user_id")
	RequestIDCtxKey = contextKey("request_id")
	TraceIDCtxKey   = contextKey("trace_id")
	EmailCtxKey     = contextKey("email")
	RoleCtxKey      = contextKey("role")
	TokenKey        = contextKey("token")
)

var (
	ErrUserNotFound     = errors.New("user not found in context")
	ErrInvalidUserID    = errors.New("invalid user ID format")
	ErrInvalidContext   = errors.New("invalid context")
	ErrMissingRequestID = errors.New("request ID not found in context")
)

type detachCtx struct {
	ctx context.Context //nolint:containedctx
}

func (d detachCtx) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (d detachCtx) Done() <-chan struct{} {
	return nil
}

func (d detachCtx) Err() error {
	return nil
}

func (d detachCtx) Value(key any) any {
	return d.ctx.Value(key)
}

func NewContextFrom(ctx context.Context) context.Context {
	return detachCtx{ctx: ctx}
}

func ExtractTraceID(ctx context.Context) string {
	traceID, _ := ctx.Value(TraceIDCtxKey).(string)
	return traceID
}

func ExtractUserID(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(UserIDCtxKey).(int64)
	if !ok {
		return 0, ErrMissingRequestID
	}
	return id, nil
}

func InjectTraceID(parentCtx context.Context, traceID string) context.Context {
	return context.WithValue(parentCtx, TraceIDCtxKey, traceID)
}

func InjectUserID(parentCtx context.Context, id int64) context.Context {
	return context.WithValue(parentCtx, UserIDCtxKey, id)
}

func ExtractEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailCtxKey).(string)
	return email, ok
}

func ExtractRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleCtxKey).(string)
	return role, ok
}

func ExtractToken(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(TokenKey).(string)
	return token, ok
}

// ExtractParentID extracts parent ID from context (parent is a user)
// Returns the user ID which represents the authenticated parent
func ExtractParentID(ctx context.Context) (int64, error) {
	id, ok := ctx.Value(UserIDCtxKey).(int64)
	if !ok {
		return 0, ErrUserNotFound
	}
	return id, nil
}
