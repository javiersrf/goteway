package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/javiersrf/goteway/services"
	"github.com/javiersrf/goteway/utils"
	"github.com/redis/go-redis/v9"
)

func NewRequestHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s\n", r.Method, r.URL.Path)

		ok, _ := services.GetFromCache(r, w, rdb)
		if ok {
			log.Printf("Cache hit for %s %s\n", r.Method, r.URL.Path)
			return
		}
		log.Printf("Cache miss for %s %s\n", r.Method, r.URL.Path)

		response, err := utils.MakeRequest(r.Method, r.URL.Path, r.Header, r.Body)
		if err != nil {
			fmt.Printf("Error making request to backend: %v\n", err)
			http.Error(w, "Error making request to backend", http.StatusInternalServerError)
			return
		}
		defer response.Body.Close()

		expirationInMinutes := 5
		if response.Header.Get("X-Cache") != "" && response.Header.Get("X-Cache-Expire") != "" {
			expiration := response.Header.Get("X-Cache-Expire")
			expirationInMinutes, err = strconv.Atoi(expiration)
			if err != nil {
				fmt.Printf("Error getting from cache: %v\n", err)
				expirationInMinutes = 240
			}
		}
		services.SaveOnCache(r, response, rdb, expirationInMinutes)

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

		w.Write(bodyBytes)

		if err != nil {
			fmt.Printf("Error setting cache: %v\n", err)
		}
	}
}
