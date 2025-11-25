package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// SessionStore provides a simple, thread-safe, in-memory storage for web sessions.
// Each session is identified by a unique token and has a defined expiry time.
type SessionStore struct {
	sessions map[string]time.Time // maps session token to its expiry time
	mu       sync.Mutex           // provides thread-safety
}

// NewSessionStore initializes and returns a new SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]time.Time),
	}
}

// CreateSession generates a cryptographically random session token, stores it with a
// 24-hour expiry, and returns the token.
func (s *SessionStore) CreateSession() string {
	b := make([]byte, 32)
	rand.Read(b) // This will panic if the OS's entropy source fails
	token := base64.URLEncoding.EncodeToString(b)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = time.Now().Add(24 * time.Hour)
	return token
}

// IsValid checks if a given session token exists in the store and has not expired.
// If the token has expired, it is removed from the store.
// It returns true if the session is valid, otherwise false.
func (s *SessionStore) IsValid(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiry, exists := s.sessions[token]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(s.sessions, token) // Clean up expired session
		return false
	}
	return true
}

// CheckCredentials verifies a given username and password against credentials
// defined in environment variables (APP_USER, APP_PASSWORD).
// For development convenience, it falls back to default credentials ("admin", "secret")
// if the environment variables are not set.
// Note: This is a simple comparison and not suitable for production without hashing.
// It returns true if the credentials are valid, otherwise false.
func CheckCredentials(user, pass string) bool {
	envUser := os.Getenv("APP_USER")
	if envUser == "" {
		envUser = "admin"
	}

	envPass := os.Getenv("APP_PASSWORD")
	if envPass == "" {
		envPass = "secret"
	}

	return user == envUser && pass == envPass
}

// Middleware creates a new HTTP middleware handler.
// This middleware protects routes by checking for a valid session token in a cookie.
// If the session is not valid, the user is redirected to the /login page.
// Static assets and the login page itself are excluded from this check.
//
// Parameters:
//   - store: The SessionStore used to validate session tokens.
//   - next: The next http.HandlerFunc in the chain to call if authentication succeeds.
//
// Returns:
//   - An http.HandlerFunc that wraps the original handler with authentication logic.
func Middleware(store *SessionStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow unauthenticated access to the login page and static assets
		if r.URL.Path == "/login" || strings.HasPrefix(r.URL.Path, "/static") {
			next(w, r)
			return
		}

		// Check for a valid session cookie
		cookie, err := r.Cookie("session_token")
		if err != nil || !store.IsValid(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// If the session is valid, proceed to the originally requested handler
		next(w, r)
	}
}

