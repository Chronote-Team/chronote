package infra

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
)

type RedisBlacklist struct {
	client *redislib.Client
}

func NewRedisBlacklist(client *redislib.Client) *RedisBlacklist {
	return &RedisBlacklist{client: client}
}

func (r *RedisBlacklist) IsBlacklisted(token string) (bool, error) {
	if r.client == nil {
		return false, nil
	}
	result, err := r.client.Exists(context.Background(), token).Result()
	return result > 0, err
}

func (r *RedisBlacklist) BlacklistTokenPair(accessToken, refreshToken string) error {
	if r.client == nil {
		return nil
	}
	if accessToken != "" {
		if err := r.client.Set(context.Background(), accessToken, "1", 0).Err(); err != nil {
			return err
		}
	}
	if refreshToken != "" {
		if err := r.client.Set(context.Background(), refreshToken, "1", 0).Err(); err != nil {
			return err
		}
	}
	return nil
}
