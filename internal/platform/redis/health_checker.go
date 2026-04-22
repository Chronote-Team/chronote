package redis

import (
	"context"
	"errors"

	redislib "github.com/redis/go-redis/v9"
)

type Checker struct {
	client *redislib.Client
}

func NewChecker(client *redislib.Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) Check(ctx context.Context) error {
	if c == nil || c.client == nil {
		return errors.New("redis not initialized")
	}
	return c.client.Ping(ctx).Err()
}
