package portfolio

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"irish-cgt-tracker/internal/currency"
	"irish-cgt-tracker/internal/models"
)

type SaleDTO struct {
    models.Sale
}

type Service struct {
	db *sql.DB
}

// Public wrapper for inventory
func (s *Service) GetInventory() ([]InventoryItem, error) {
	return s.getAvailableInventory()
}

func (s *Service) GetAllSales() ([]SaleDTO, error) {
	rows, err := s.db.Query("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sales []SaleDTO
	for rows.Next() {
		var item SaleDTO
		if err := rows.Scan(&item.ID, &item.Date, &item.Quantity, &item.PriceCents, &item.ECBRate, &item.IsSettled); err != nil {
			return nil, err
		}
		sales = append(sales, item)
	}
	return sales, nil
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

func (s *Service) getSale(id string) (*models.Sale, error) {
	var sale models.Sale
	row := s.db.QueryRow("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales WHERE id = ?", id)
	if err := row.Scan(&sale.ID, &sale.Date, &sale.Quantity, &sale.PriceCents, &sale.ECBRate, &sale.IsSettled); err != nil {
		return nil, err
	}
	return &sale, nil
}

// getAvailableInventory fetches all vests and subtracts shares already used in other sales.
// Ordered by Date ASC to support FIFO.
func (s *Service) getAvailableInventory() ([]InventoryItem, error) {
	// Query: Select Vest details AND the sum of quantities used in sale_lots
	query := `
		SELECT
			v.id, v.date, v.symbol, v.quantity, v.strike_price_cents, v.ecb_rate,
			COALESCE(SUM(sl.quantity), 0) as used_qty
		FROM vests v
		LEFT JOIN sale_lots sl ON v.id = sl.vest_id
		GROUP BY v.id
		ORDER BY v.date ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []InventoryItem
	for rows.Next() {
		var item InventoryItem
		var usedQty int64
		if err := rows.Scan(&item.ID, &item.Date, &item.Symbol, &item.Quantity, &item.StrikePriceCents, &item.ECBRate, &usedQty); err != nil {
			return nil, err
		}
		item.RemainingQty = item.Quantity - usedQty

		// Only add to inventory if there are shares left
		if item.RemainingQty > 0 {
			inventory = append(inventory, item)
		}
	}
	return inventory, nil
}

func (s *Service) saveLot(saleID, vestID string, qty int64) error {
	_, err := s.db.Exec("INSERT INTO sale_lots (sale_id, vest_id, quantity) VALUES (?, ?, ?)", saleID, vestID, qty)
	return err
}

func (s *Service) markSaleSettled(saleID string) error {
	_, err := s.db.Exec("UPDATE sales SET is_settled = 1 WHERE id = ?", saleID)
	return err
}

