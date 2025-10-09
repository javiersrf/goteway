package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheValue struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

func SaveOnCache(r *http.Request, resp *http.Response, rdb *redis.Client, expirationMinutes int) error {
	authHeader := r.Header.Get("Authorization")

	cacheKey := fmt.Sprintf("%s-%s-%s", r.Method, r.URL.Path, authHeader)

	encodedCacheKey := base64.URLEncoding.EncodeToString([]byte(cacheKey))

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	cacheValue := CacheValue{
		Body:    string(bodyBytes),
		Headers: headers,
	}

	encodedContent, err := json.Marshal(cacheValue)
	if err != nil {
		return fmt.Errorf("error encoding cache value: %w", err)
	}

	ctx := r.Context()
	err = rdb.Set(ctx, encodedCacheKey, encodedContent, time.Duration(expirationMinutes)*time.Minute).Err()

	if err != nil {
		return fmt.Errorf("error saving to redis: %w", err)
	}

	return nil
}

func GetFromCache(r *http.Request, w http.ResponseWriter, rdb *redis.Client) (bool, error) {
	authHeader := r.Header.Get("Authorization")
	cacheKey := fmt.Sprintf("%s-%s-%s", r.Method, r.URL.Path, authHeader)
	encodedCacheKey := base64.URLEncoding.EncodeToString([]byte(cacheKey))

	ctx := r.Context()
	val, err := rdb.Get(ctx, encodedCacheKey).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("error getting from redis: %w", err)
	}

	var cached CacheValue
	if err := json.Unmarshal([]byte(val), &cached); err != nil {
		return false, fmt.Errorf("error decoding cache value: %w", err)
	}

	for k, v := range cached.Headers {
		w.Header().Set(k, v)
	}

	_, err = w.Write([]byte(cached.Body))
	if err != nil {
		return false, fmt.Errorf("error writing cached response: %w", err)
	}

	return true, nil
}
