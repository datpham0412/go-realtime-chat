package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/tinrab/graphql-realtime-chat/server"
)

type config struct {
	RedisURL string `envconfig:"REDIS_URL"`
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found")
	}

	var cfg config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	// Debug environment variables
	log.Printf("[DEBUG] Environment variables:")
	log.Printf("GitHub Client ID: %s", os.Getenv("GITHUB_CLIENT_ID"))
	log.Printf("GitHub Client Secret length: %d", len(os.Getenv("GITHUB_CLIENT_SECRET")))
	log.Printf("Redis URL: %s", redisURL)

	s, err := server.NewServer(redisURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] Server created successfully")

	err = s.Serve(8080)
	if err != nil {
		log.Fatal(err)
	}
}
