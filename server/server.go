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
	githubapi "github.com/google/go-github/v32/github"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"golang.org/x/oauth2"
	githubauth "golang.org/x/oauth2/github"
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
		var baseURL string

		// Get the origin from the request header
		origin := r.Header.Get("Origin")
		if origin != "" {
			baseURL = origin
		} else {
			// Fallback to host-based URL construction
			if r.TLS != nil || os.Getenv("NODE_ENV") == "production" {
				baseURL = "https://" + r.Host
			} else {
				baseURL = "http://" + r.Host
			}
		}

		// For development, always use the backend URL
		if os.Getenv("NODE_ENV") != "production" {
			baseURL = "http://localhost:8080"
		}

		callbackURL = baseURL + "/auth/github/callback"
		fmt.Printf("Using callback URL: %s\n", callbackURL)

		// Create the OAuth config
		oauthConfig := &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  callbackURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: githubauth.Endpoint,
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
		var baseURL string

		// Determine the base URL from the request
		if r.TLS != nil || os.Getenv("NODE_ENV") == "production" {
			baseURL = "https://" + r.Host
		} else {
			baseURL = "http://localhost:8080"
		}

		callbackURL = baseURL + "/auth/github/callback"

		oauthConfig := &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  callbackURL,
			Scopes: []string{
				"user:email",
				"read:user",
			},
			Endpoint: githubauth.Endpoint,
		}

		token, err := oauthConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		// Get user info from GitHub
		client := githubapi.NewClient(oauthConfig.Client(r.Context(), token))
		user, _, err := client.Users.Get(r.Context(), "")
		if err != nil {
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		// Set the user cookie with the GitHub username
		cookie := &http.Cookie{
			Name:     "user_id",
			Value:    *user.Login,
			Path:     "/",
			HttpOnly: false,
			Secure:   r.TLS != nil || os.Getenv("NODE_ENV") == "production",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   86400,
			Domain:   r.Host,
		}
		http.SetCookie(w, cookie)

		// Log cookie setting
		fmt.Printf("Setting cookie: %+v\n", cookie)

		// Determine redirect URL
		redirectURL := "/"
		if os.Getenv("NODE_ENV") == "production" {
			redirectURL = "https://" + r.Host
		} else {
			redirectURL = "http://localhost:3000"
		}

		// Add a small delay to ensure cookie is set
		time.Sleep(100 * time.Millisecond)

		// Redirect with JavaScript to ensure cookie is properly set
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<html>
			<body>
				<script>
					document.cookie = "user_id=%s;path=/;max-age=86400;%s";
					window.location.href = "%s";
				</script>
			</body>
			</html>
		`, *user.Login,
			func() string {
				if r.TLS != nil || os.Getenv("NODE_ENV") == "production" {
					return "secure;samesite=lax"
				}
				return "samesite=lax"
			}(),
			redirectURL)
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
