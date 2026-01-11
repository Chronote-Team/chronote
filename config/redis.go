package config

import (
	"chronote/global"
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedis initializes the Redis client connection
func InitRedis() {
	addr := fmt.Sprintf("%s:%s", AppConfig.Redis.Host, AppConfig.Redis.Port)

	global.RedisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: AppConfig.Redis.Password,
		DB:       AppConfig.Redis.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := global.RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established successfully")
}
