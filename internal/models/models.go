package models

import "time"

// Vest represents a single stock vesting event, which is treated as an
// acquisition for Capital Gains Tax purposes.
type Vest struct {
	ID string `json:"id"` // Unique identifier (UUID) for the vesting event.
	// Date of the vest in "YYYY-MM-DD" format.
	Date string `json:"date"`
	// Symbol is the stock ticker, e.g., "GOOGL".
	Symbol string `json:"symbol"`
	// Quantity is the number of shares that vested.
	Quantity int64 `json:"quantity"`
	// StrikePriceCents is the market price of a single share in USD cents at the time of vesting.
	StrikePriceCents int64 `json:"strike_price_cents"`
	// ECBRate is the ECB reference exchange rate (EUR per 1 USD) on the vesting date.
	ECBRate float64 `json:"ecb_rate"`
}

// Sale represents a single stock sale event, treated as a disposal for CGT.
type Sale struct {
	ID string `json:"id"` // Unique identifier (UUID) for the sale event.
	// Date of the sale in "YYYY-MM-DD" format.
	Date string `json:"date"`
	// Quantity is the total number of shares sold in this event.
	Quantity int64 `json:"quantity"`
	// PriceCents is the price of a single share in USD cents at the time of sale.
	PriceCents int64 `json:"price_cents"`
	// ECBRate is the ECB reference exchange rate (EUR per 1 USD) on the sale date.
	ECBRate float64 `json:"ecb_rate"`
	// IsSettled is a flag indicating whether the CGT implications for this sale
	// have been calculated and accounted for.
	IsSettled bool `json:"is_settled"`
}

// SaleLot represents a component of a Sale, linking a specific number of shares
// from that sale back to their original Vest. This is necessary for implementing
// the FIFO (First-In, First-Out) accounting rule.
type SaleLot struct {
	// SaleID is the foreign key referencing the parent Sale.
	SaleID string `json:"sale_id"`
	// VestID is the foreign key referencing the source Vest.
	VestID string `json:"vest_id"`
	// Quantity is the number of shares from the specified Vest that were
	// disposed of in this specific Sale.
	Quantity int64 `json:"quantity"`
}

// ParseDate is a utility function to parse a date string in "YYYY-MM-DD" format
// into a time.Time object.
//
// Parameters:
//   - dateStr: The date string to parse.
//
// Returns:
//   - A time.Time object representing the parsed date.
//   - An error if the string is not in the expected format.
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
