package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/frahmantamala/jadiles/internal"
	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	ctxUserIDKey = "user_id"
	ctxEmailKey  = "email"
	ctxRoleKey   = "role"
	ctxTokenKey  = "token"
)

type TokenClaims struct {
	UserID int64
	Email  string
	Role   string
	JTI    string
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTAuthentication struct {
	accessSecret         []byte
	refreshSecret        []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	issuer               string
}

func NewJWTAuthentication(config internal.HTTPServerConfig) (*JWTAuthentication, error) {
	accessSecret, err := config.GetJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to decode jwt secret: %w", err)
	}

	refreshSecret, err := config.GetRefreshTokenSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to decode refresh token secret: %w", err)
	}

	return &JWTAuthentication{
		accessSecret:         accessSecret,
		refreshSecret:        refreshSecret,
		accessTokenDuration:  config.AuthConfig.AccessTokenDuration,
		refreshTokenDuration: config.AuthConfig.RefreshTokenDuration,
		issuer:               config.AuthConfig.Issuer,
	}, nil
}

func (ja *JWTAuthentication) GenerateAccessToken(ctx context.Context, userID int64, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(ja.accessTokenDuration)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    ja.issuer,
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ja.accessSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken creates a new JWT refresh token
func (ja *JWTAuthentication) GenerateRefreshToken(ctx context.Context, userID int64, email, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(ja.refreshTokenDuration)

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    ja.issuer,
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(ja.refreshSecret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (ja *JWTAuthentication) ParseAccessToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	return ja.parseToken(ctx, tokenString, ja.accessSecret)
}

func (ja *JWTAuthentication) ParseRefreshToken(ctx context.Context, tokenString string) (*TokenClaims, error) {
	return ja.parseToken(ctx, tokenString, ja.refreshSecret)
}

func (ja *JWTAuthentication) parseToken(ctx context.Context, tokenString string, secret []byte) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		JTI:    claims.ID,
	}, nil
}

func (ja *JWTAuthentication) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ja.extractTokenFromHeader(r)
		if token == "" {
			ja.handleAuthError(w, r, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := ja.ParseAccessToken(r.Context(), token)
		if err != nil {
			ja.handleAuthError(w, r, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, ctxEmailKey, claims.Email)
		ctx = context.WithValue(ctx, ctxRoleKey, claims.Role)
		ctx = context.WithValue(ctx, ctxTokenKey, token)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (ja *JWTAuthentication) OptionalAuthenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := ja.extractTokenFromHeader(r)
		if token != "" {
			claims, err := ja.ParseAccessToken(r.Context(), token)
			if err == nil {
				ctx := context.WithValue(r.Context(), ctxUserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, ctxEmailKey, claims.Email)
				ctx = context.WithValue(ctx, ctxRoleKey, claims.Role)
				ctx = context.WithValue(ctx, ctxTokenKey, token)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (ja *JWTAuthentication) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(ctxRoleKey).(string)
			if !ok {
				ja.handleAuthError(w, r, "unauthorized", http.StatusUnauthorized)
				return
			}

			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				ja.handleAuthError(w, r, "insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (ja *JWTAuthentication) AccessTokenDuration() time.Duration {
	return ja.accessTokenDuration
}

func (ja *JWTAuthentication) RefreshTokenDuration() time.Duration {
	return ja.refreshTokenDuration
}

func (ja *JWTAuthentication) extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

func (ja *JWTAuthentication) handleAuthError(w http.ResponseWriter, r *http.Request, message string, statusCode int) {
	resp := &v1.DefaultErrorResponse{}
	resp.Error.Message = message
	render.Status(r, statusCode)
	render.JSON(w, r, resp)
}
