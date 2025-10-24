package auth

import (
	"context"
	"fmt"
	"time"

	goRedis "github.com/redis/go-redis/v9"
)

type RedisTokenStorage struct {
	client goRedis.UniversalClient
}

func NewRedisTokenStorage(client goRedis.UniversalClient) *RedisTokenStorage {
	return &RedisTokenStorage{
		client: client,
	}
}

func (r *RedisTokenStorage) StoreToken(ctx context.Context, userID int64, token string, expiry time.Duration) error {
	key := r.generateTokenKey(userID, token)
	err := r.client.Set(ctx, key, "valid", expiry).Err()
	if err != nil {
		return fmt.Errorf("failed to store token: %w", err)
	}
	return nil
}

func (r *RedisTokenStorage) IsTokenValid(ctx context.Context, userID int64, token string) (bool, error) {
	key := r.generateTokenKey(userID, token)
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == goRedis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("failed to check token validity: %w", err)
	}
	return result == "valid", nil
}

func (r *RedisTokenStorage) InvalidateToken(ctx context.Context, userID int64, token string) error {
	key := r.generateTokenKey(userID, token)
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate token: %w", err)
	}
	return nil
}

func (r *RedisTokenStorage) InvalidateAllUserTokens(ctx context.Context, userID int64) error {
	pattern := r.generateUserTokenPattern(userID)

	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan user tokens: %w", err)
	}

	if len(keys) > 0 {
		err := r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete user tokens: %w", err)
		}
	}

	return nil
}

func (r *RedisTokenStorage) generateTokenKey(userID int64, token string) string {
	return fmt.Sprintf("token:user:%d:%s", userID, token)
}

func (r *RedisTokenStorage) generateUserTokenPattern(userID int64) string {
	return fmt.Sprintf("token:user:%d:*", userID)
}
