package portfolio

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"irish-cgt-tracker/internal/models"
)

func TestSettleSale_SimpleFIFO(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewService(db)

	saleID := uuid.New().String()
	vestID := uuid.New().String()

	sale := &models.Sale{
		ID:         saleID,
		Date:       "2024-02-01",
		Quantity:   100,
		PriceCents: 15000, // $150.00
		ECBRate:    0.9,   // 1 USD = 0.9 EUR
		IsSettled:  false,
	}

	vest := &InventoryItem{
		Vest: models.Vest{
			ID:               vestID,
			Date:             "2024-01-01",
			Symbol:           "TEST",
			Quantity:         200,
			StrikePriceCents: 10000, // $100.00
			ECBRate:          0.8,   // 1 USD = 0.8 EUR
		},
		RemainingQty: 200,
	}

	// 1. Expect the getSale query
	mock.ExpectQuery("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales WHERE id = ?").
		WithArgs(saleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "quantity", "price_cents", "ecb_rate", "is_settled"}).
			AddRow(sale.ID, sale.Date, sale.Quantity, sale.PriceCents, sale.ECBRate, sale.IsSettled))

	// 2. Expect the getAvailableInventory query
	mock.ExpectQuery("SELECT v.id, v.date, v.symbol, v.quantity, v.strike_price_cents, v.ecb_rate, COALESCE(SUM(sl.quantity), 0) as used_qty FROM vests v LEFT JOIN sale_lots sl ON v.id = sl.vest_id GROUP BY v.id ORDER BY v.date ASC").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "symbol", "quantity", "strike_price_cents", "ecb_rate", "used_qty"}).
			AddRow(vest.Vest.ID, vest.Vest.Date, vest.Vest.Symbol, vest.Vest.Quantity, vest.Vest.StrikePriceCents, vest.Vest.ECBRate, 0))

	// 3. Expect the saveLot insert
	mock.ExpectExec("INSERT INTO sale_lots (sale_id, vest_id, quantity) VALUES (?, ?, ?)").
		WithArgs(saleID, vestID, int64(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. Expect the markSaleSettled update
	mock.ExpectExec("UPDATE sales SET is_settled = 1 WHERE id = ?").
		WithArgs(saleID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = s.SettleSale(saleID)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSettleSale_InsufficientInventory(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewService(db)
	saleID := uuid.New().String()

	sale := &models.Sale{
		ID:         saleID,
		Date:       "2024-02-01",
		Quantity:   100,
		PriceCents: 15000,
		ECBRate:    0.9,
		IsSettled:  false,
	}

	// 1. Expect the getSale query
	mock.ExpectQuery("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales WHERE id = ?").
		WithArgs(saleID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "quantity", "price_cents", "ecb_rate", "is_settled"}).
			AddRow(sale.ID, sale.Date, sale.Quantity, sale.PriceCents, sale.ECBRate, sale.IsSettled))

	// 2. Expect the getAvailableInventory query to return an empty inventory
	mock.ExpectQuery("SELECT v.id, v.date, v.symbol, v.quantity, v.strike_price_cents, v.ecb_rate, COALESCE(SUM(sl.quantity), 0) as used_qty FROM vests v LEFT JOIN sale_lots sl ON v.id = sl.vest_id GROUP BY v.id ORDER BY v.date ASC").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "symbol", "quantity", "strike_price_cents", "ecb_rate", "used_qty"}))

	err = s.SettleSale(saleID)
	if err == nil {
		t.Error("expected an error due to insufficient inventory, but got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
