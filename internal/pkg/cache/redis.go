package cache

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

type RedisCache struct {
	ctx    context.Context
	client *redis.Client
}

func InitRedisCache(ctx context.Context, host string, port int) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	// check connection by setting test value
	err := rdb.Set(ctx, "key", "value", 0).Err()

	return &RedisCache{
		ctx:    ctx,
		client: rdb,
	}, err
}

func (c *RedisCache) Add(key int, expiration int64) error {
	return c.client.Set(c.ctx, strconv.Itoa(key), "value", time.Duration(expiration*1e9)*time.Second).Err()
}

func (c *RedisCache) Get(key int) (bool, error) {
	val, err := c.client.Get(c.ctx, strconv.Itoa(key)).Result()
	return val != "", err
}

func (c *RedisCache) Delete(key int) {
	c.client.Del(c.ctx, strconv.Itoa(key))
}
