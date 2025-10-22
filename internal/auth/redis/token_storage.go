package redis

import (
	"context"
	"fmt"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

// TokenStorage handles JWT token storage and invalidation in Redis
type TokenStorage struct {
	client goRedis.UniversalClient
}

// NewTokenStorage creates a new token storage instance
func NewTokenStorage(client goRedis.UniversalClient) *TokenStorage {
	return &TokenStorage{
		client: client,
	}
}

// StoreToken stores a JWT token in Redis with expiration
func (ts *TokenStorage) StoreToken(ctx context.Context, userID int64, token string, expiry time.Duration) error {
	key := fmt.Sprintf("auth:token:%d:%s", userID, token)

	err := ts.client.Set(ctx, key, "valid", expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to store token in redis: %w", err)
	}

	return nil
}

// IsTokenValid checks if a token is valid (exists in Redis and not blacklisted)
func (ts *TokenStorage) IsTokenValid(ctx context.Context, userID int64, token string) (bool, error) {
	key := fmt.Sprintf("auth:token:%d:%s", userID, token)

	val, err := ts.client.Get(ctx, key).Result()
	if err == goRedis.Nil {
		// Token not found or expired
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check token in redis: %w", err)
	}

	return val == "valid", nil
}

// InvalidateToken invalidates a specific token (logout)
func (ts *TokenStorage) InvalidateToken(ctx context.Context, userID int64, token string) error {
	key := fmt.Sprintf("auth:token:%d:%s", userID, token)

	err := ts.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate token in redis: %w", err)
	}

	return nil
}

// InvalidateAllUserTokens invalidates all tokens for a user (e.g., on password change)
func (ts *TokenStorage) InvalidateAllUserTokens(ctx context.Context, userID int64) error {
	pattern := fmt.Sprintf("auth:token:%d:*", userID)

	// Scan and delete all matching keys
	iter := ts.client.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := ts.client.Del(ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete token %s: %w", iter.Val(), err)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan tokens: %w", err)
	}

	return nil
}
