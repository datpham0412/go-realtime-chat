package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	oauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/github/callback",
		Scopes: []string{
			"user:email",
			"read:user",
		},
		Endpoint: github.Endpoint,
	}
)

type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

func init() {
	log.Printf("Initializing OAuth config...")
	log.Printf("Redirect URL: %s", "http://localhost:8080/auth/github/callback")
}

func (s *Server) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("GitHub login handler called")
	log.Printf("Method: %s", r.Method)
	log.Printf("URL: %s", r.URL.String())

	url := oauthConfig.AuthCodeURL("state")
	log.Printf("Generated GitHub OAuth URL: %s", url)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	log.Printf("Redirected to GitHub")
}

func (s *Server) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for the callback
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	code := r.URL.Query().Get("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	client := oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Set cookie with proper settings
	http.SetCookie(w, &http.Cookie{
		Name:     "user_id",
		Value:    githubUser.Login,
		Path:     "/",
		Domain:   "localhost",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: false,     // Changed to false so JavaScript can read it
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to frontend with success parameter
	http.Redirect(w, r, "http://localhost:3000?login=success", http.StatusTemporaryRedirect)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	cookie, err := r.Cookie("user_id")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Not logged in"}`))
		return
	}

	userKey := "user:" + cookie.Value
	userData, err := s.redis.Get(r.Context(), userKey).Result()
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "User not found"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(userData))
}
