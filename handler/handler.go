package handler

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/javiersrf/goteway/services"
	"github.com/javiersrf/goteway/utils"
	"github.com/redis/go-redis/v9"
)

func NewRequestHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[START] Received request: %s %s", r.Method, r.URL.String())
		log.Printf("[HEADERS] %v", r.Header)

		ok, _ := services.GetFromCache(r, w, rdb)
		if ok {
			log.Printf("[CACHE HIT] %s %s (duration: %v)", r.Method, r.URL.Path, time.Since(start))
			return
		}
		log.Printf("[CACHE MISS] %s %s", r.Method, r.URL.Path)

		queryParams := make(map[string]string)
		for key, values := range r.URL.Query() {
			if len(values) > 0 {
				queryParams[key] = values[0]
			}
		}

		response, err := utils.MakeRequest(r.Method, r.URL.Path, r.Header, queryParams, r.Body)
		if err != nil {
			log.Printf("[ERROR] Backend request failed: %v", err)
			http.Error(w, "Error making request to backend", http.StatusInternalServerError)
			return
		}
		defer func() {
			if cerr := response.Body.Close(); cerr != nil {
				log.Printf("[WARN] Failed to close response body: %v", cerr)
			}
		}()

		log.Printf("[BACKEND RESPONSE] Status: %d | Headers: %v", response.StatusCode, response.Header)

		expirationInMinutes := 5
		log.Printf("[CACHE HEADERS] Checking for X-Cache headers")

		log.Printf("[DEBUG] Response Headers: %v", response.Header)
		xCache := response.Header.Get("X-Cache")
		xCacheExpire := response.Header.Get("X-Cache-Expire")

		if xCache != "" && xCacheExpire != "" {
			log.Printf("[CACHE HEADER DETECTED] X-Cache: %s | X-Cache-Expire: %s", xCache, xCacheExpire)
			expirationInMinutes, err = strconv.Atoi(xCacheExpire)
			if err != nil {
				log.Printf("[ERROR] Invalid X-Cache-Expire value '%s': %v. Using default 240 minutes.", xCacheExpire, err)
				expirationInMinutes = 240
			}
			log.Printf("[CACHE SAVE] Attempting to cache response for %d minutes", expirationInMinutes)
			if err := services.SaveOnCache(r, response, rdb, expirationInMinutes); err != nil {
				log.Printf("[ERROR] Failed to save on cache: %v", err)
			} else {
				log.Printf("[CACHE SAVE SUCCESS] %s %s cached for %d minutes", r.Method, r.URL.Path, expirationInMinutes)
			}
		} else {
			log.Printf("[CACHE SKIP] No X-Cache headers found, skipping cache storage")
		}

		for key, values := range response.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		w.WriteHeader(response.StatusCode)

		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			log.Printf("[ERROR] Reading response body failed: %v", err)
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		w.Write(bodyBytes)
		log.Printf("[RESPONSE SENT] %s %s | Status: %d | Duration: %v | Body size: %d bytes",
			r.Method, r.URL.Path, response.StatusCode, time.Since(start), len(bodyBytes))
	}
}
