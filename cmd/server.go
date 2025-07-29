package main

import (
	"go-redis/config"
	"go-redis/internal"
	"go-redis/internal/infrastructure/input-ports/http"
	"os"
)

func main() {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		panic("Failed to create data directory: " + err.Error())
	}

	serviceConfig := config.NewConfig().GetConfig()
	useCases := internal.NewAdapterUseCaseBridge(serviceConfig)
	server := http.NewServer(useCases, serviceConfig)
	server.Start()
}
