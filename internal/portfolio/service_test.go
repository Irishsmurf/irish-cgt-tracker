package portfolio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"irish-cgt-tracker/internal/currency"
)

func TestAddVest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"rates":{"EUR":0.9}}`))
	}))
	defer server.Close()

	originalBaseURL := currency.BaseURL
	currency.BaseURL = server.URL
	defer func() { currency.BaseURL = originalBaseURL }()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewService(db)

	mock.ExpectExec("INSERT INTO vests").
		WithArgs(sqlmock.AnyArg(), "2024-01-01", "TEST", int64(100), int64(10000), 0.9).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = s.AddVest("2024-01-01", "TEST", 100, 10000)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestAddSale(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"rates":{"EUR":0.9}}`))
	}))
	defer server.Close()

	originalBaseURL := currency.BaseURL
	currency.BaseURL = server.URL
	defer func() { currency.BaseURL = originalBaseURL }()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := NewService(db)

	mock.ExpectExec("INSERT INTO sales").
		WithArgs(sqlmock.AnyArg(), "2024-02-01", int64(50), int64(12000), 0.9, false).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = s.AddSale("2024-02-01", 50, 12000)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
