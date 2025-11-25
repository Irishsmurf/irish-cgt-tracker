package portfolio

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"irish-cgt-tracker/internal/currency"
	"irish-cgt-tracker/internal/models"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// AddVest records a new vesting event.
// It automatically fetches the ECB rate for the given date.
func (s *Service) AddVest(date string, symbol string, qty int64, strikePriceCents int64) (*models.Vest, error) {
	// 1. Fetch Exchange Rate
	rate, err := currency.FetchUSDToEUR(date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate for %s: %w", date, err)
	}

	// 2. Create Model
	vest := &models.Vest{
		ID:               uuid.New().String(),
		Date:             date,
		Symbol:           symbol,
		Quantity:         qty,
		StrikePriceCents: strikePriceCents,
		ECBRate:          rate,
	}

	// 3. Insert into DB
	query := `INSERT INTO vests (id, date, symbol, quantity, strike_price_cents, ecb_rate) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, vest.ID, vest.Date, vest.Symbol, vest.Quantity, vest.StrikePriceCents, vest.ECBRate)
	if err != nil {
		return nil, fmt.Errorf("failed to insert vest: %w", err)
	}

	log.Printf("Vest recorded: %d shares of %s on %s @ %.4f EUR/USD", qty, symbol, date, rate)
	return vest, nil
}

// AddSale records a new sale event.
// It automatically fetches the ECB rate for the given date.
func (s *Service) AddSale(date string, qty int64, priceCents int64) (*models.Sale, error) {
	// 1. Fetch Exchange Rate
	rate, err := currency.FetchUSDToEUR(date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate for %s: %w", date, err)
	}

	// 2. Create Model
	sale := &models.Sale{
		ID:         uuid.New().String(),
		Date:       date,
		Quantity:   qty,
		PriceCents: priceCents,
		ECBRate:    rate,
		IsSettled:  false,
	}

	// 3. Insert into DB
	query := `INSERT INTO sales (id, date, quantity, price_cents, ecb_rate, is_settled) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, sale.ID, sale.Date, sale.Quantity, sale.PriceCents, sale.ECBRate, sale.IsSettled)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sale: %w", err)
	}

	log.Printf("Sale recorded: %d shares on %s @ %.4f EUR/USD", qty, date, rate)
	return sale, nil
}
