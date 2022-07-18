package main

import (
	"context"
	"fmt"
	"github.com/timurkash/pow/internal/client"
	"github.com/timurkash/pow/internal/pkg/config"
	"log"
)

func main() {
	log.Println("start client")

	configInst, err := config.Load("config/config.json")
	if err != nil {
		log.Println("error load config:", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)

	address := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)

	err = client.Run(ctx, address)
	if err != nil {
		log.Fatalln("client error:", err)
	}
}
