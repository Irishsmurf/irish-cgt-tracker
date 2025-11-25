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

// SessionStore is a simple in-memory store for active sessions
type SessionStore struct {
	sessions map[string]time.Time // Token -> Expiry
	mu       sync.Mutex
}

func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]time.Time),
	}
}

// GenerateToken creates a random session ID
func (s *SessionStore) CreateSession() string {
	b := make([]byte, 32)
	rand.Read(b)
	token := base64.URLEncoding.EncodeToString(b)
	
	s.mu.Lock()
	defer s.mu.Unlock()
	// Session valid for 24 hours
	s.sessions[token] = time.Now().Add(24 * time.Hour)
	return token
}

func (s *SessionStore) IsValid(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	expiry, exists := s.sessions[token]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(s.sessions, token)
		return false
	}
	return true
}

// Simple credential check (In prod, use bcrypt!)
func CheckCredentials(user, pass string) bool {
	// Default to "admin" / "secret" if env vars not set
	envUser := os.Getenv("APP_USER")
	if envUser == "" { envUser = "admin" }
	
	envPass := os.Getenv("APP_PASSWORD")
	if envPass == "" { envPass = "secret" }

	return user == envUser && pass == envPass
}

// Middleware to protect routes
func Middleware(store *SessionStore, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Allow static assets and login page
		if r.URL.Path == "/login" || strings.HasPrefix(r.URL.Path, "/static") {
			next(w, r)
			return
		}

		// 2. Check Cookie
		cookie, err := r.Cookie("session_token")
		if err != nil || !store.IsValid(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// 3. Authorized
		next(w, r)
	}
}

