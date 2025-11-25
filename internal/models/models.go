package models

import "time"

// Vest represents a vesting event (Acquisition)
type Vest struct {
	ID               string  `json:"id"`                 // UUID
	Date             string  `json:"date"`               // YYYY-MM-DD
	Symbol           string  `json:"symbol"`             // e.g., "GOOG"
	Quantity         int64   `json:"quantity"`           // Number of shares
	StrikePriceCents int64   `json:"strike_price_cents"` // Price in USD cents
	ECBRate          float64 `json:"ecb_rate"`           // EUR per 1 USD
}

// Sale represents a disposal event
type Sale struct {
	ID         string  `json:"id"`          // UUID
	Date       string  `json:"date"`        // YYYY-MM-DD
	Quantity   int64   `json:"quantity"`    // Number of shares sold
	PriceCents int64   `json:"price_cents"` // Sale price in USD cents
	ECBRate    float64 `json:"ecb_rate"`    // EUR per 1 USD
	IsSettled  bool    `json:"is_settled"`  // If tax has been calculated
}

// SaleLot maps specific sold shares to their original vest
type SaleLot struct {
	SaleID   string `json:"sale_id"`
	VestID   string `json:"vest_id"`
	Quantity int64  `json:"quantity"`
}

// Helper method to parse ISO dates if needed later
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// SettledSale represents a completed sale with its full tax breakdown.
// This is the model for our exportable spreadsheet view.
type SettledSale struct {
	SaleDate           string  // from Sale
	Ticker             string  // from Vest
	NumShares          int64   // from SaleLot
	SalePriceUSD       int64   // from Sale
	GainLossUSD        int64   // Calculated: (SalePriceUSD - VestPriceUSD) * NumShares
	BookValueUSD       int64   // Calculated: VestPriceUSD * NumShares
	ExchangeRateAtVest float64 // from Vest
	GrossProceedUSD    int64   // Calculated: SalePriceUSD * NumShares
	VestingValueUSD    int64   // from Vest
	ExchangeRateAtSale float64 // from Sale
	EuroSaleEUR        int64   // Calculated: GrossProceedUSD * ExchangeRateAtSale
	EuroGainEUR        int64   // Calculated: EuroSaleEUR - (BookValueUSD * ExchangeRateAtVest)
	CGTTaxDueEUR       int64   // Calculated: EuroGainEUR * 0.33
	Completed          string  // Always "Y"
	NetProceedsEUR     int64   // Calculated: EuroSaleEUR - CGTTaxDueEUR
	Type               string  // Always "FIFO"
}
