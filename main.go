package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/javiersrf/goteway/cache"
	"github.com/javiersrf/goteway/handler"
	"github.com/javiersrf/goteway/utils"
)

type Response struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

var ctx = context.Background()

func main() {

	rdb := cache.InitializeRedis(ctx)
	defer rdb.Close()

	utils.InitClient()
	requestHandler := handler.NewRequestHandler(rdb)
	serverPort := "8080"
	fmt.Printf("Starting server on port %s\n", serverPort)
	envServerPort := os.Getenv("SERVER_PORT")
	if envServerPort != "" {
		serverPort = envServerPort
	}
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), requestHandler); err != nil {
		panic(err)
	}

}
