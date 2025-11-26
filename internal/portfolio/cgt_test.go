package portfolio

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSettleSale_Simple(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewService(db)

	// --- Mocking ---
	// 1. GetSale
	mock.ExpectQuery("SELECT id, date, quantity, price_cents, ecb_rate, is_settled FROM sales WHERE id = ?").
		WithArgs("sale1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "quantity", "price_cents", "ecb_rate", "is_settled"}).
			AddRow("sale1", "2024-02-01", 50, 15000, 0.9, false))

	// 2. GetInventory
	mock.ExpectQuery("SELECT v.id, v.date, v.symbol, v.quantity, v.strike_price_cents, v.ecb_rate, COALESCE(SUM(sl.quantity), 0) as used_qty FROM vests v LEFT JOIN sale_lots sl ON v.id = sl.vest_id GROUP BY v.id ORDER BY v.date ASC").
		WillReturnRows(sqlmock.NewRows([]string{"id", "date", "symbol", "quantity", "strike_price_cents", "ecb_rate", "used_qty"}).
			AddRow("vest1", "2024-01-01", "TEST", 100, 10000, 0.8, 0))

	// 3. SaveLot
	mock.ExpectExec("INSERT INTO sale_lots (sale_id, vest_id, quantity) VALUES (?, ?, ?)").
		WithArgs("sale1", "vest1", 50).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 4. insertSettledSale
	mock.ExpectExec("INSERT INTO settled_sales ( sale_date, ticker, num_shares, sale_price_usd, gain_loss_usd, book_value_usd, exchange_rate_at_vest, gross_proceed_usd, vesting_value_usd, exchange_rate_at_sale, euro_sale_eur, euro_gain_eur, cgt_tax_due_eur, completed, net_proceeds_eur, type ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)").
		WithArgs("2024-02-01", "TEST", int64(50), int64(750000), int64(250000), int64(500000), 0.8, int64(750000), int64(500000), 0.9, int64(675000), int64(275000), int64(90750), "Y", int64(584250), "FIFO").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. markSaleSettled
	mock.ExpectExec("UPDATE sales SET is_settled = 1 WHERE id = ?").
		WithArgs("sale1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// --- Execution ---
	err = s.SettleSale("sale1")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// --- Verification ---
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
