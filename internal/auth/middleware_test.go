package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestSessionStore(t *testing.T) {
	store := NewSessionStore()

	// Test session creation
	token := store.CreateSession()
	if token == "" {
		t.Fatal("expected a session token, but got an empty string")
	}

	// Test session validation
	if !store.IsValid(token) {
		t.Error("newly created session token is not valid")
	}

	// Test expired session
	store.mu.Lock()
	store.sessions[token] = time.Now().Add(-1 * time.Hour)
	store.mu.Unlock()
	if store.IsValid(token) {
		t.Error("expired session token should not be valid")
	}

	// Test non-existent session
	if store.IsValid("non-existent-token") {
		t.Error("non-existent session token should not be valid")
	}
}

func TestCheckCredentials(t *testing.T) {
	// Test with default credentials
	if !CheckCredentials("admin", "secret") {
		t.Error("expected default credentials to be valid")
	}

	// Test with environment variables
	os.Setenv("APP_USER", "testuser")
	os.Setenv("APP_PASSWORD", "testpass")
	if !CheckCredentials("testuser", "testpass") {
		t.Error("expected credentials from environment variables to be valid")
	}
	os.Unsetenv("APP_USER")
	os.Unsetenv("APP_PASSWORD")

	// Test invalid credentials
	if CheckCredentials("wronguser", "wrongpass") {
		t.Error("expected invalid credentials to be invalid")
	}
}

func TestMiddleware(t *testing.T) {
	store := NewSessionStore()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := Middleware(store, handler)

	// Test without session cookie
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, but got %d", http.StatusSeeOther, rr.Code)
	}

	// Test with invalid session cookie
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-token"})
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, but got %d", http.StatusSeeOther, rr.Code)
	}

	// Test with valid session cookie
	token := store.CreateSession()
	req, _ = http.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rr = httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, but got %d", http.StatusOK, rr.Code)
	}

	// Test static assets and login page should be accessible
	paths := []string{"/login", "/static/style.css"}
	for _, path := range paths {
		req, _ = http.NewRequest("GET", path, nil)
		rr = httptest.NewRecorder()
		middleware.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("expected status %d for path %s, but got %d", http.StatusOK, path, rr.Code)
		}
	}
}
