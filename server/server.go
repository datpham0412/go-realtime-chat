package server

import (
	"fmt"
	"net/http"
	"time"

	"context"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type WebsocketInitFunc func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error)

type Server struct {
	redis    *redis.Client
	mux      *http.ServeMux
	handler  http.Handler
	upgrader websocket.Upgrader
}

func NewServer(redisURL string) (*Server, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	server := &Server{
		redis: client,
		mux:   http.NewServeMux(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/auth/github", s.handleGitHubLogin)
	s.mux.HandleFunc("/auth/github/callback", s.handleGitHubCallback)

	resolver := &Resolver{redis: s.redis}
	srv := handler.New(NewExecutableSchema(Config{Resolvers: resolver}))

	srv.AddTransport(transport.Websocket{
		Upgrader:              s.upgrader,
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	s.mux.Handle("/graphql", srv)
	s.mux.Handle("/", playground.Handler("GraphQL playground", "/graphql"))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "ws://localhost:3000", "ws://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "Sec-WebSocket-Protocol", "apollographql-client-name", "apollographql-client-version"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		Debug:            true,
	})

	s.handler = corsHandler.Handler(s.mux)
}

func (s *Server) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, s.handler)
}
