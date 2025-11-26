package importer

import (
	"strings"
	"testing"
)

func TestParseVestCSV(t *testing.T) {
	csvData := `Vest Date,Order Number,Plan,Type,Status,Price,Quantity,Net Cash Proceeds,Net Share Proceeds,Tax Payment Method
25-Nov-2025,RB9995EE17,GSU Class C,Release,Staged,$318.47,14.094,$0.00,6.752,Fractional Shares`

	reader := strings.NewReader(csvData)
	vests, err := ParseVestCSV(reader)
	if err != nil {
		t.Fatalf("ParseVestCSV failed: %v", err)
	}

	if len(vests) != 1 {
		t.Fatalf("Expected 1 vest, got %d", len(vests))
	}

	vest := vests[0]
	if vest.Date != "2025-11-25" {
		t.Errorf("Expected date 2025-11-25, got %s", vest.Date)
	}
	if vest.Quantity != 14.094 {
		t.Errorf("Expected quantity 14.094, got %f", vest.Quantity)
	}
	if vest.StrikePriceCents != 31847 {
		t.Errorf("Expected price 31847, got %d", vest.StrikePriceCents)
	}
}

func TestParseSaleCSV(t *testing.T) {
	csvData := `Execution Date,Order Number,Plan,Type,Order Status,Price,Quantity,Net Amount,Net Share Proceeds,Tax Payment Method
18-Mar-2025,WBC8F81C195-1EE,Cash,Sale,Complete,$1.00,-179.720,$179.72,0,N/A`

	reader := strings.NewReader(csvData)
	sales, err := ParseSaleCSV(reader)
	if err != nil {
		t.Fatalf("ParseSaleCSV failed: %v", err)
	}

	if len(sales) != 1 {
		t.Fatalf("Expected 1 sale, got %d", len(sales))
	}

	sale := sales[0]
	if sale.Date != "2025-03-18" {
		t.Errorf("Expected date 2025-03-18, got %s", sale.Date)
	}
	if sale.Quantity != 179.720 {
		t.Errorf("Expected quantity 179.720, got %f", sale.Quantity)
	}
	if sale.PriceCents != 100 {
		t.Errorf("Expected price 100, got %d", sale.PriceCents)
	}
}
