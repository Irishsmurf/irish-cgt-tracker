package server

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"irish-cgt-tracker/internal/db"
	"irish-cgt-tracker/internal/portfolio"
)

func TestHandleImport(t *testing.T) {
	db, cleanup := db.NewTestDB(t)
	defer cleanup()

	svc := portfolio.NewService(db)
	server := NewServer(svc, false, "../../web/templates")

	// Test GET
	req, _ := http.NewRequest("GET", "/import", nil)
	rr := httptest.NewRecorder()
	server.handleImport(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Test POST - Vests
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("csvFile", "vests.csv")
	io.WriteString(part, `Vest Date,Order Number,Plan,Type,Status,Price,Quantity,Net Cash Proceeds,Net Share Proceeds,Tax Payment Method
25-Nov-2025,RB9995EE17,GSU Class C,Release,Staged,$318.47,14.094,$0.00,6.752,Fractional Shares`)
	writer.WriteField("importType", "vests")
	writer.WriteField("symbol", "GOOGL")
	writer.Close()

	req, _ = http.NewRequest("POST", "/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr = httptest.NewRecorder()
	server.handleImport(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	inventory, _ := svc.GetInventory()
	if len(inventory) != 1 {
		t.Fatalf("expected 1 vest, got %d", len(inventory))
	}
	if inventory[0].Quantity != 14.094 {
		t.Errorf("expected quantity 14.094, got %f", inventory[0].Quantity)
	}
	if inventory[0].Symbol != "GOOGL" {
		t.Errorf("expected symbol GOOGL, got %s", inventory[0].Symbol)
	}

	// Test POST - Sales
	body = &bytes.Buffer{}
	writer = multipart.NewWriter(body)
	part, _ = writer.CreateFormFile("csvFile", "sales.csv")
	io.WriteString(part, `Execution Date,Order Number,Plan,Type,Order Status,Price,Quantity,Net Amount,Net Share Proceeds,Tax Payment Method
18-Mar-2025,WBC8F81C195-1EE,Cash,Sale,Complete,$1.00,-179.720,$179.72,0,N/A`)
	writer.WriteField("importType", "sales")
	writer.Close()

	req, _ = http.NewRequest("POST", "/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr = httptest.NewRecorder()
	server.handleImport(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", rr.Code)
	}

	sales, _ := svc.GetAllSales()
	if len(sales) != 1 {
		t.Fatalf("expected 1 sale, got %d", len(sales))
	}
	if sales[0].Quantity != 179.720 {
		t.Errorf("expected quantity 179.720, got %f", sales[0].Quantity)
	}
	if sales[0].PriceCents != 100 {
		t.Errorf("expected price 100, got %d", sales[0].PriceCents)
	}
}
