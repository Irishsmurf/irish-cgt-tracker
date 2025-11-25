package currency

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchUSDToEUR(t *testing.T) {
	// Test server that mocks the Frankfurter API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/2024-01-01" {
			fmt.Fprintln(w, `{"rates":{"EUR":0.8}}`)
		} else if r.URL.Path == "/2024-01-02" {
			w.WriteHeader(http.StatusNotFound)
		} else if r.URL.Path == "/2024-01-03" {
			fmt.Fprintln(w, `{]`) // Invalid JSON
		} else if r.URL.Path == "/2024-01-04" {
			fmt.Fprintln(w, `{"rates":{}}`) // EUR rate missing
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	// Override the BaseURL to use the test server
	originalBaseURL := BaseURL
	BaseURL = server.URL
	defer func() { BaseURL = originalBaseURL }()

	// Test successful fetch
	rate, err := FetchUSDToEUR("2024-01-01")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if rate != 0.8 {
		t.Errorf("expected rate 0.8, but got %f", rate)
	}

	// Test fallback on 404
	rate, err = FetchUSDToEUR("2024-01-02")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if rate != 0.8 {
		t.Errorf("expected rate 0.8, but got %f", rate)
	}

	// Test invalid date format
	_, err = FetchUSDToEUR("invalid-date")
	if err == nil {
		t.Error("expected an error for invalid date format, but got nil")
	}

	// Test invalid JSON response
	_, err = FetchUSDToEUR("2024-01-03")
	if err == nil {
		t.Error("expected an error for invalid JSON, but got nil")
	}

	// Test missing EUR rate in response
	_, err = FetchUSDToEUR("2024-01-04")
	if err == nil {
		t.Error("expected an error for missing EUR rate, but got nil")
	}

	// Test API error
	_, err = FetchUSDToEUR("2024-01-05")
	if err == nil {
		t.Error("expected an error for API error, but got nil")
	}
}
