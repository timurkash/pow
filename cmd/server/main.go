package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/timurkash/pow/internal/pkg/cache"
	"github.com/timurkash/pow/internal/pkg/clock"
	"github.com/timurkash/pow/internal/pkg/config"
	"github.com/timurkash/pow/internal/server"
)

func main() {
	log.Println("start server")

	configInst, err := config.Load("config/config.json")
	if err != nil {
		log.Fatalln("error load config:", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)
	ctx = context.WithValue(ctx, "clock", clock.SystemClock{})

	cacheInst, err := cache.InitRedisCache(ctx, configInst.CacheHost, configInst.CachePort)
	if err != nil {
		log.Fatalln("error init cache:", err)
		return
	}
	ctx = context.WithValue(ctx, "cache", cacheInst)

	rand.Seed(time.Now().UnixNano())

	serverAddress := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)
	err = server.Run(ctx, serverAddress)
	if err != nil {
		log.Fatalln("server error:", err)
	}
}
