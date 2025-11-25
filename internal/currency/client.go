package currency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	// Using the standard public instance. 
	// If you have a specific internal mirror, change this URL.
	BaseURL = "https://api.frankfurter.app" 
	MaxRetries = 5
)

// rateResponse matches the JSON structure from Frankfurter
type rateResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// FetchUSDToEUR retrieves the ECB reference rate for a given date.
// If the date is a weekend/holiday, it backtracks to find the nearest previous rate.
func FetchUSDToEUR(dateStr string) (float64, error) {
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < MaxRetries; i++ {
		// Format the date for the current attempt
		currentDateStr := targetDate.Format("2006-01-02")
		url := fmt.Sprintf("%s/%s?from=USD&to=EUR", BaseURL, currentDateStr)

		resp, err := client.Get(url)
		if err != nil {
			// Network error? Wait a sec and retry
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		// If 404 (Not Found), it's likely a weekend/holiday in the API
		if resp.StatusCode == http.StatusNotFound {
			targetDate = targetDate.AddDate(0, 0, -1) // Go back 1 day
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("API error: received status %d", resp.StatusCode)
		}

		// Parse the response
		var result rateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return 0, fmt.Errorf("failed to decode JSON: %v", err)
		}

		// Extract EUR rate
		rate, exists := result.Rates["EUR"]
		if !exists {
			// If for some reason EUR is missing, treat as failure and backtrack
			targetDate = targetDate.AddDate(0, 0, -1)
			continue
		}

		// Success!
		// Optionally log if we had to backtrack
		if currentDateStr != dateStr {
			fmt.Printf("Notice: No rate for %s. Used rate from %s: %.4f\n", dateStr, currentDateStr, rate)
		}

		return rate, nil
	}

	return 0, fmt.Errorf("could not find an ECB rate for %s within %d days", dateStr, MaxRetries)
}
