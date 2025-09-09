package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func InitializeRedis(ctx context.Context) *redis.Client {
	redisURL := os.Getenv("REDISADDR")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(fmt.Sprintf("invalid redis url: %v", err))
	}
	rdb := redis.NewClient(opt)

	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	fmt.Println("Connected:", pong)
	return rdb
}

func SetCache(ctx context.Context, rdb *redis.Client, key string, value string, expiration time.Duration) error {
	cacheKey := fmt.Sprintf("item:%s", key)
	err := rdb.Set(ctx, cacheKey, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func GetCache(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	val, err := rdb.Get(ctx, fmt.Sprintf("item:%s", key)).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", err
	}
	return val, nil
}

func DeleteCache(ctx context.Context, rdb *redis.Client, key string) error {
	_, err := rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}
