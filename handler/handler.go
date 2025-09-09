package handler

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/javiersrf/goteway/cache"
	"github.com/javiersrf/goteway/utils"
	"github.com/redis/go-redis/v9"
)

func NewRequestHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cacheKey := fmt.Sprint(r.Method + "-" + r.URL.Path)
		encodedCacheKey := base64.URLEncoding.EncodeToString([]byte(cacheKey))

		val, err := cache.GetCache(r.Context(), rdb, encodedCacheKey)
		if err != nil {
			fmt.Printf("Error getting from cache: %v\n", err)
		}
		if val != "" {
			fmt.Println("Serving from cache")
			decodedContent, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				fmt.Printf("Error getting from cache: %v\n", err)
			}
			w.Write([]byte(decodedContent))
			return
		}
		response, err := utils.MakeRequest(r.Method, r.URL.Path, r.Header)
		if err != nil {
			fmt.Printf("Error making request to backend: %v\n", err)
			http.Error(w, "Error making request to backend", http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(response.StatusCode)
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		if response.Header.Get("X-Cache") != "" && response.Header.Get("X-Cache-Expire") != "" {
			expiration := response.Header.Get("X-Cache-Expire")
			var expirationInt int
			expirationInt, err := strconv.Atoi(expiration)
			if err != nil {
				fmt.Printf("Error getting from cache: %v\n", err)
				expirationInt = 240
			}

			encodedContent := base64.StdEncoding.EncodeToString([]byte(bodyBytes))

			cache.SetCache(r.Context(), rdb, encodedCacheKey, encodedContent, time.Duration(expirationInt)*time.Minute)
		}
		w.Write(bodyBytes)

		if err != nil {
			fmt.Printf("Error setting cache: %v\n", err)
		}
	}
}
