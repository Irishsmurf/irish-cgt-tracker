package currency

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	// BaseURL is the endpoint for the Frankfurter.app API.
	// This can be overridden for testing or to use a self-hosted instance.
	BaseURL = "https://api.frankfurter.app"
	// MaxRetries defines the number of days to look back when searching for a valid
	// exchange rate if the initial date is a holiday or weekend.
	MaxRetries = 5
)

// rateResponse defines the structure of the JSON response from the Frankfurter API.
type rateResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// FetchUSDToEUR queries the Frankfurter API to get the historical USD to EUR
// exchange rate for a specific date, as published by the European Central Bank (ECB).
//
// The function is designed to be resilient to non-trading days (weekends, holidays).
// If the API returns a 404 Not Found for the requested date, it automatically
// retries by requesting the rate for the previous day. This process is repeated up
// to MaxRetries times.
//
// Parameters:
//   - dateStr: The date for which to fetch the rate, in "YYYY-MM-DD" format.
//
// Returns:
//   - A float64 representing the EUR equivalent of 1 USD for the given date.
//   - An error if the date format is invalid, the API is unreachable after retries,
//     or if a rate cannot be found within the retry limit.
func FetchUSDToEUR(dateStr string) (float64, error) {
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for i := 0; i < MaxRetries; i++ {
		currentDateStr := targetDate.Format("2006-01-02")
		url := fmt.Sprintf("%s/%s?from=USD&to=EUR", BaseURL, currentDateStr)

		resp, err := client.Get(url)
		if err != nil {
			// On network error, wait briefly before the next attempt.
			time.Sleep(500 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()

		// A 404 status indicates the requested date is a non-trading day.
		// We backtrack one day and try again.
		if resp.StatusCode == http.StatusNotFound {
			targetDate = targetDate.AddDate(0, 0, -1) // Go back 1 day
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("API error: received status %d", resp.StatusCode)
		}

		var result rateResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return 0, fmt.Errorf("failed to decode JSON: %v", err)
		}

		rate, exists := result.Rates["EUR"]
		if !exists {
			// If the EUR rate is unexpectedly missing, backtrack and retry.
			targetDate = targetDate.AddDate(0, 0, -1)
			continue
		}

		// Success. If we had to backtrack, log a notice for transparency.
		if currentDateStr != dateStr {
			fmt.Printf("Notice: No rate for %s. Used rate from %s: %.4f\n", dateStr, currentDateStr, rate)
		}
		return rate, nil
	}

	return 0, fmt.Errorf("could not find an ECB rate for %s within %d days", dateStr, MaxRetries)
}
