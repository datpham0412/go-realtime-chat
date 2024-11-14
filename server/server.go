package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"context"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
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
	// Setup GitHub OAuth routes
	s.mux.HandleFunc("/auth/github", func(w http.ResponseWriter, r *http.Request) {
		// Get the environment-specific callback URL
		var callbackURL string
		if os.Getenv("NODE_ENV") == "production" {
			callbackURL = "https://go-realtime-chat.fly.dev/auth/github/callback"
		} else {
			callbackURL = "http://localhost:8080/auth/github/callback"
		}

		// Create the OAuth config
		oauthConfig := &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  callbackURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: github.Endpoint,
		}

		// Generate the authorization URL
		url := oauthConfig.AuthCodeURL("state")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	s.mux.HandleFunc("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}

		var callbackURL string
		if os.Getenv("NODE_ENV") == "production" {
			callbackURL = "https://go-realtime-chat.fly.dev/auth/github/callback"
		} else {
			callbackURL = "http://localhost:8080/auth/github/callback"
		}

		oauthConfig := &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  callbackURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: github.Endpoint,
		}

		token, err := oauthConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		// Set the user cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "user_id",
			Value:    token.AccessToken,
			Path:     "/",
			HttpOnly: true,
			Secure:   os.Getenv("NODE_ENV") == "production",
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to the frontend
		if os.Getenv("NODE_ENV") == "production" {
			http.Redirect(w, r, "https://go-realtime-chat.fly.dev", http.StatusTemporaryRedirect)
		} else {
			http.Redirect(w, r, "http://localhost:3000", http.StatusTemporaryRedirect)
		}
	})

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

	// Only show playground in development
	if os.Getenv("NODE_ENV") != "production" {
		s.mux.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://go-realtime-chat.fly.dev",
			"ws://localhost:3000",
			"ws://localhost:8080",
			"wss://go-realtime-chat.fly.dev",
		},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Authorization", "Sec-WebSocket-Protocol", "apollographql-client-name", "apollographql-client-version"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		Debug:            false,
	})

	// Serve static files in production
	s.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// First, try to serve static files
		staticPath := "./static"
		fs := http.FileServer(http.Dir(staticPath))

		// Check if the path is /graphql or starts with /auth
		if r.URL.Path == "/graphql" || strings.HasPrefix(r.URL.Path, "/auth/") {
			s.handler.ServeHTTP(w, r)
			return
		}

		// Check if file exists
		path := filepath.Join(staticPath, r.URL.Path)
		_, err := os.Stat(path)

		if os.IsNotExist(err) {
			// File doesn't exist, serve index.html for SPA routing
			http.ServeFile(w, r, filepath.Join(staticPath, "index.html"))
			return
		}

		// File exists, serve it
		fs.ServeHTTP(w, r)
	})

	s.handler = corsHandler.Handler(s.mux)
}

func (s *Server) Serve(port int) error {
	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, s.handler)
}
