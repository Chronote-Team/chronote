package services

import (
	"chronote/global"
	"context"
	"fmt"
	"time"
)

const (
	// Token blacklist key prefix in Redis
	tokenBlacklistPrefix = "token:blacklist:"

	// AccessToken TTL: 7200 seconds (2 hours)
	AccessTokenBlacklistTTL = 7200 * time.Second

	// RefreshToken TTL: 1814400 seconds (21 days)
	RefreshTokenBlacklistTTL = 1814400 * time.Second
)

type TokenBlacklistService struct{}

// AddToBlacklist adds a token to the Redis blacklist with specified TTL
func (s *TokenBlacklistService) AddToBlacklist(ctx context.Context, token string, ttl time.Duration) error {
	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, token)
	return global.RedisClient.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted checks if a token exists in the blacklist
func (s *TokenBlacklistService) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("%s%s", tokenBlacklistPrefix, token)
	result, err := global.RedisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// BlacklistTokenPair adds both access and refresh tokens to the blacklist
func (s *TokenBlacklistService) BlacklistTokenPair(ctx context.Context, accessToken, refreshToken string) error {
	// Add access token to blacklist
	if err := s.AddToBlacklist(ctx, accessToken, AccessTokenBlacklistTTL); err != nil {
		return fmt.Errorf("failed to blacklist access token: %w", err)
	}

	// Add refresh token to blacklist
	if err := s.AddToBlacklist(ctx, refreshToken, RefreshTokenBlacklistTTL); err != nil {
		return fmt.Errorf("failed to blacklist refresh token: %w", err)
	}

	return nil
}
