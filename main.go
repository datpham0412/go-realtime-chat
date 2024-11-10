package main

import (
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/graphql-realtime-chat/server"
)

type config struct {
	RedisURL string `envconfig:"REDIS_URL"`
}

func main() {
	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379" // fallback for local development
	}

	s, err := server.NewGraphQLServer(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	err = s.Serve("/graphql", 8080)
	if err != nil {
		log.Fatal(err)
	}
}
