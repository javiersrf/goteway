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
	addresRedis := os.Getenv("REDISADDR")
	if addresRedis == "" {
		addresRedis = "localhost:6379"
	}
	passwordRedis := os.Getenv("REDISPASSWORD")
	if passwordRedis == "" {
		passwordRedis = ""
	}
	dbRedis := os.Getenv("REDISDB")
	var db int
	if dbRedis == "" {
		db = 0
	} else {
		fmt.Sscanf(dbRedis, "%d", &db)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addresRedis,
		Password: passwordRedis,
		DB:       db,
	})

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
