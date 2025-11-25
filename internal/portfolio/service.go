package portfolio

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"irish-cgt-tracker/internal/currency"
	"irish-cgt-tracker/internal/models"
)

// SaleDTO (Data Transfer Object) is a simple wrapper around the models.Sale struct.
// It's used to transfer sale data, particularly for presentation layers, without
// necessarily exposing the full internal model.
type SaleDTO struct {
	models.Sale
}

// Service provides methods for managing and calculating portfolio data.
// It encapsulates the core business logic and database interactions.
type Service struct {
	db *sql.DB
}

// NewService creates and returns a new Service instance.
//
// Parameters:
//   - db: An active sql.DB connection pool for database operations.
//
// Returns:
//   - A pointer to the newly created Service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetInventory provides a public interface to the getAvailableInventory method.
// It returns a list of all vested shares that still have a remaining quantity unsold.
func (s *Service) GetInventory() ([]InventoryItem, error) {
	return s.getAvailableInventory()
}

// GetAllSales retrieves all sale records from the database, ordered by date descending.
//
// Returns:
//   - A slice of SaleDTO objects.
//   - An error if the database query fails.
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

// AddVest creates and stores a new stock vesting event.
// It automatically fetches the required ECB USD/EUR exchange rate for the vesting date
// before persisting the record to the database.
//
// Parameters:
//   - date: The vesting date in "YYYY-MM-DD" format.
//   - symbol: The stock ticker symbol.
//   - qty: The number of shares that vested.
//   - strikePriceCents: The market price per share in USD cents at vest time.
//
// Returns:
//   - A pointer to the newly created models.Vest object.
//   - An error if the exchange rate cannot be fetched or the database insertion fails.
func (s *Service) AddVest(date string, symbol string, qty int64, strikePriceCents int64) (*models.Vest, error) {
	rate, err := currency.FetchUSDToEUR(date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate for %s: %w", date, err)
	}

	vest := &models.Vest{
		ID:               uuid.New().String(),
		Date:             date,
		Symbol:           symbol,
		Quantity:         qty,
		StrikePriceCents: strikePriceCents,
		ECBRate:          rate,
	}

	query := `INSERT INTO vests (id, date, symbol, quantity, strike_price_cents, ecb_rate) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, vest.ID, vest.Date, vest.Symbol, vest.Quantity, vest.StrikePriceCents, vest.ECBRate)
	if err != nil {
		return nil, fmt.Errorf("failed to insert vest: %w", err)
	}

	log.Printf("Vest recorded: %d shares of %s on %s @ %.4f EUR/USD", qty, symbol, date, rate)
	return vest, nil
}

// AddSale creates and stores a new stock sale event.
// Similar to AddVest, it automatically fetches the ECB USD/EUR exchange rate
// for the sale date before persisting the record. The new sale is initially
// marked as unsettled.
//
// Parameters:
//   - date: The sale date in "YYYY-MM-DD" format.
//   - qty: The number of shares sold.
//   - priceCents: The sale price per share in USD cents.
//
// Returns:
//   - A pointer to the newly created models.Sale object.
//   - An error if the exchange rate cannot be fetched or the database insertion fails.
func (s *Service) AddSale(date string, qty int64, priceCents int64) (*models.Sale, error) {
	rate, err := currency.FetchUSDToEUR(date)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch exchange rate for %s: %w", date, err)
	}

	sale := &models.Sale{
		ID:         uuid.New().String(),
		Date:       date,
		Quantity:   qty,
		PriceCents: priceCents,
		ECBRate:    rate,
		IsSettled:  false,
	}

	query := `INSERT INTO sales (id, date, quantity, price_cents, ecb_rate, is_settled) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = s.db.Exec(query, sale.ID, sale.Date, sale.Quantity, sale.PriceCents, sale.ECBRate, sale.IsSettled)
	if err != nil {
		return nil, fmt.Errorf("failed to insert sale: %w", err)
	}

	log.Printf("Sale recorded: %d shares on %s @ %.4f EUR/USD", qty, date, rate)
	return sale, nil
}

// getSale retrieves a single sale record by its ID. This is an internal helper function.
func (s *Service) getSale(id string) (*models.Sale, error) {
	var sale models.Sale
	row := s.db.QueryRow("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales WHERE id = ?", id)
	if err := row.Scan(&sale.ID, &sale.Date, &sale.Quantity, &sale.PriceCents, &sale.ECBRate, &sale.IsSettled); err != nil {
		return nil, err
	}
	return &sale, nil
}

// getAvailableInventory calculates the current inventory of unsold shares.
// It queries all vests and joins with the sale_lots table to determine how many
// shares from each vest have been sold. It returns a list of vests that still
// have a positive remaining quantity, ordered by date (oldest first) to support FIFO.
func (s *Service) getAvailableInventory() ([]InventoryItem, error) {
	query := `
		SELECT
			v.id, v.date, v.symbol, v.quantity, v.strike_price_cents, v.ecb_rate,
			COALESCE(SUM(sl.quantity), 0) as used_qty
		FROM vests v
		LEFT JOIN sale_lots sl ON v.id = sl.vest_id
		GROUP BY v.id
		ORDER BY v.date ASC`
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

		if item.RemainingQty > 0 {
			inventory = append(inventory, item)
		}
	}
	return inventory, nil
}

// saveLot records the link between a sale and a vest for a specific quantity of shares.
// This is an internal helper function called by the SettleSale calculator.
func (s *Service) saveLot(saleID, vestID string, qty int64) error {
	_, err := s.db.Exec("INSERT INTO sale_lots (sale_id, vest_id, quantity) VALUES (?, ?, ?)", saleID, vestID, qty)
	return err
}

// markSaleSettled updates the status of a sale to 'settled' in the database.
// This is an internal helper function called by the SettleSale calculator.
func (s *Service) markSaleSettled(saleID string) error {
	_, err := s.db.Exec("UPDATE sales SET is_settled = 1 WHERE id = ?", saleID)
	return err
}

