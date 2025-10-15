package internal

import (
	"errors"
	"fmt"
	"net/http"

	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/render"
)

// Common application errors
var (
	// Authentication & Authorization errors
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden access")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrExpiredToken       = errors.New("token has expired")
	ErrTokenNotFound      = errors.New("token not found")
	ErrInvalidClaims      = errors.New("invalid token claims")
	ErrInvalidSignature   = errors.New("invalid token signature")

	// Validation errors
	ErrValidationFailed     = errors.New("validation failed")
	ErrInvalidInput         = errors.New("invalid input data")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidFormat        = errors.New("invalid data format")
	ErrValidation           = errors.New("validation error")

	// Resource errors
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	ErrConflict      = errors.New("resource conflict")
	ErrGone          = errors.New("resource no longer available")

	// Database errors
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseConstraint = errors.New("database constraint violation")
	ErrTransactionFailed  = errors.New("database transaction failed")

	// Business logic errors
	ErrBusinessRule     = errors.New("business rule violation")
	ErrInvalidOperation = errors.New("invalid operation")
	ErrOperationFailed  = errors.New("operation failed")
	ErrInvalidState     = errors.New("invalid state")

	// System errors
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrTimeout            = errors.New("operation timeout")
	ErrRateLimitExceeded  = errors.New("rate limit exceeded")

	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenIsBlocked = errors.New("refresh token is blocked")

	ErrOptimisticLock  = errors.New("optimistic locking error")
	ErrUniqueViolation = errors.New("unique violation error")

	ErrDataNotFound              = errors.New("data not found")
	ErrForbiddenToUseApplication = errors.New("user is forbidden to use this application. please contact your administrator")
)

// AppError represents a structured application error
type AppError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"-"`
	Err        error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

func (e *AppError) WithDetail(key string, value interface{}) *AppError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error constructors
func NewAppError(code, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

type ValidationError struct {
	Msg string
}

func NewValidationError(msg string) *ValidationError {
	return &ValidationError{Msg: msg}
}

func (v *ValidationError) Error() string {
	return v.Msg
}

func NewNotFoundError(resource string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound, ErrNotFound)
}

func NewConflictError(message string, err error) *AppError {
	return NewAppError("CONFLICT", message, http.StatusConflict, err)
}

func NewUnauthorizedError(message string) *AppError {
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized, ErrUnauthorized)
}

func NewForbiddenError(message string) *AppError {
	return NewAppError("FORBIDDEN", message, http.StatusForbidden, ErrForbidden)
}

func NewInternalServerError(err error) *AppError {
	return NewAppError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError, err)
}

func NewBusinessRuleError(message string, err error) *AppError {
	return NewAppError("BUSINESS_RULE_VIOLATION", message, http.StatusUnprocessableEntity, err)
}

// Error helpers
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

func GetStatusCode(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// Validation error details
type ValidationErrorDetail struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

func NewValidationErrorWithDetails(details []ValidationErrorDetail) *ValidationError {
	return &ValidationError{Msg: "Validation failed"}
}

// Error wrapping utilities
func WrapError(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func WrapAppError(err error, code, message string, statusCode int) *AppError {
	return NewAppError(code, message, statusCode, err)
}

// Error logging context
type ErrorContext struct {
	UserID    string                 `json:"user_id,omitempty"`
	RequestID string                 `json:"request_id,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Resource  string                 `json:"resource,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

func NewErrorContext() *ErrorContext {
	return &ErrorContext{
		Details: make(map[string]interface{}),
	}
}

func (ec *ErrorContext) WithUserID(userID string) *ErrorContext {
	ec.UserID = userID
	return ec
}

func (ec *ErrorContext) WithRequestID(requestID string) *ErrorContext {
	ec.RequestID = requestID
	return ec
}

func (ec *ErrorContext) WithOperation(operation string) *ErrorContext {
	ec.Operation = operation
	return ec
}

func (ec *ErrorContext) WithResource(resource string) *ErrorContext {
	ec.Resource = resource
	return ec
}

func (ec *ErrorContext) WithDetail(key string, value interface{}) *ErrorContext {
	ec.Details[key] = value
	return ec
}

func HandleEndpointError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	resp := &v1.DefaultErrorResponse{}
	httpCode := http.StatusInternalServerError
	resp.Error.Message = err.Error()

	switch {
	case errors.Is(err, ErrRefreshTokenNotFound), errors.Is(err, ErrRefreshTokenIsBlocked):
		httpCode = http.StatusUnauthorized
	case errors.Is(err, ErrOptimisticLock):
		httpCode = http.StatusBadRequest
		resp.Error.Message = "something has changed. please reload and try again."
	case errors.Is(err, ErrValidation),
		errors.Is(err, ErrDataNotFound):
		httpCode = http.StatusBadRequest
	case errors.Is(err, ErrForbiddenToUseApplication),
		errors.Is(err, ErrForbidden):
		httpCode = http.StatusForbidden
	}

	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) {
		httpCode = http.StatusBadRequest
	}

	render.Status(r, httpCode)
	render.JSON(w, r, resp)
}
