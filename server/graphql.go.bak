//go:generate go run github.com/99designs/gqlgen
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/segmentio/ksuid"
)

type contextKey string

const (
	userContextKey = contextKey("user")
)

type graphQLServer struct {
	redisClient     *redis.Client
	messageChannels map[string]chan *Message
	userChannels    map[string]chan string
	mutex           sync.Mutex
}

func NewGraphQLServer(redisURL string) (*graphQLServer, error) {
	log.Printf("Initializing GraphQL server with Redis URL: %s", redisURL)

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("Failed to parse Redis URL: %v", err)
		return nil, err
	}

	client := redis.NewClient(opts)

	// Test Redis connection with retries
	connected := false
	maxRetries := 5
	retryCount := 0

	for !connected && retryCount < maxRetries {
		_, err := client.Ping().Result()
		if err != nil {
			log.Printf("Redis connection attempt %d failed: %v", retryCount+1, err)
			retryCount++
			time.Sleep(2 * time.Second)
			continue
		}
		connected = true
		log.Printf("Successfully connected to Redis")
		break
	}

	if !connected {
		return nil, fmt.Errorf("failed to connect to Redis after %d attempts", maxRetries)
	}

	server := &graphQLServer{
		redisClient:     client,
		messageChannels: map[string]chan *Message{},
		userChannels:    map[string]chan string{},
		mutex:           sync.Mutex{},
	}

	log.Printf("GraphQL server initialized successfully")
	return server, nil
}

func (s *graphQLServer) Serve(route string, port int) error {
	mux := http.NewServeMux()

	// Serve static files (frontend)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Check Redis connection
		_, err := s.redisClient.Ping().Result()
		if err != nil {
			log.Printf("Health check failed: Redis error: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// GraphQL endpoint
	mux.Handle(
		route,
		handler.GraphQL(NewExecutableSchema(Config{Resolvers: s}),
			handler.WebsocketUpgrader(websocket.Upgrader{
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
				HandshakeTimeout: 20 * time.Second,
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			}),
			handler.WebsocketKeepAliveDuration(20*time.Second),
		),
	)
	mux.Handle("/playground", handler.Playground("GraphQL", route))

	// Setup CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	// Create server with specific configuration
	server := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		Handler: corsHandler,
	}

	// Log server start
	log.Printf("Server starting on http://0.0.0.0:%d", port)

	// Start server
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server error: %v", err)
		return err
	}

	return nil
}

func (s *graphQLServer) PostMessage(ctx context.Context, user string, text string) (*Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	// Create message
	m := &Message{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now().UTC(),
		Text:      text,
		User:      user,
	}
	mj, _ := json.Marshal(m)
	if err := s.redisClient.LPush("messages", mj).Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	// Notify new message
	s.mutex.Lock()
	for _, ch := range s.messageChannels {
		ch <- m
	}
	s.mutex.Unlock()
	return m, nil
}

func (s *graphQLServer) Messages(ctx context.Context) ([]*Message, error) {
	cmd := s.redisClient.LRange("messages", 0, -1)
	if cmd.Err() != nil {
		log.Println(cmd.Err())
		return nil, cmd.Err()
	}
	res, err := cmd.Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	messages := []*Message{}
	for _, mj := range res {
		m := &Message{}
		err = json.Unmarshal([]byte(mj), &m)
		messages = append(messages, m)
	}
	return messages, nil
}

func (s *graphQLServer) Users(ctx context.Context) ([]string, error) {
	cmd := s.redisClient.SMembers("users")
	if cmd.Err() != nil {
		log.Println(cmd.Err())
		return nil, cmd.Err()
	}
	res, err := cmd.Result()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return res, nil
}

func (s *graphQLServer) MessagePosted(ctx context.Context, user string) (<-chan *Message, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	// Create new channel for request
	messages := make(chan *Message, 1)
	s.mutex.Lock()
	s.messageChannels[user] = messages
	s.mutex.Unlock()

	// Delete channel when done
	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.messageChannels, user)
		s.mutex.Unlock()
	}()

	return messages, nil
}

func (s *graphQLServer) UserJoined(ctx context.Context, user string) (<-chan string, error) {
	err := s.createUser(user)
	if err != nil {
		return nil, err
	}

	// Create new channel for request
	users := make(chan string, 1)
	s.mutex.Lock()
	s.userChannels[user] = users
	s.mutex.Unlock()

	// Delete channel when done
	go func() {
		<-ctx.Done()
		s.mutex.Lock()
		delete(s.userChannels, user)
		s.mutex.Unlock()
	}()

	return users, nil
}

func (s *graphQLServer) createUser(user string) error {
	// Upsert user
	if err := s.redisClient.SAdd("users", user).Err(); err != nil {
		return err
	}
	// Notify new user joined
	s.mutex.Lock()
	for _, ch := range s.userChannels {
		ch <- user
	}
	s.mutex.Unlock()
	return nil
}

func (s *graphQLServer) Mutation() MutationResolver {
	return s
}

func (s *graphQLServer) Query() QueryResolver {
	return s
}

func (s *graphQLServer) Subscription() SubscriptionResolver {
	return s
}
