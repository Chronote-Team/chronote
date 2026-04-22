package redis

import (
	"fmt"

	platformconfig "chronote-refactor/internal/platform/config"

	redislib "github.com/redis/go-redis/v9"
)

func NewClient(cfg *platformconfig.Config) *redislib.Client {
	return redislib.NewClient(&redislib.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}
