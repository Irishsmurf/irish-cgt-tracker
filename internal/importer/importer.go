package importer

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"
	"time"

	"irish-cgt-tracker/internal/models"
)

// ParseVestCSV parses a CSV file of RSU releases and returns a slice of Vest objects.
func ParseVestCSV(r io.Reader) ([]models.Vest, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var vests []models.Vest
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Vest Date,Order Number,Plan,Type,Status,Price,Quantity,Net Cash Proceeds,Net Share Proceeds,Tax Payment Method
		// 25-Nov-2025,RB9995EE17,GSU Class C,Release,Staged,$318.47,14.094,$0.00,6.752,Fractional Shares
		vestDate, err := time.Parse("02-Jan-2006", record[0])
		if err != nil {
			return nil, err
		}

		priceStr := strings.TrimPrefix(record[5], "$")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}
		priceCents := int64(price * 100)

		quantity, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			return nil, err
		}

		vests = append(vests, models.Vest{
			Date:             vestDate.Format("2006-01-02"),
			Quantity:         quantity,
			StrikePriceCents: priceCents,
		})
	}

	return vests, nil
}

// ParseSaleCSV parses a CSV file of sales and returns a slice of Sale objects.
func ParseSaleCSV(r io.Reader) ([]models.Sale, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var sales []models.Sale
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Execution Date,Order Number,Plan,Type,Order Status,Price,Quantity,Net Amount,Net Share Proceeds,Tax Payment Method
		// 18-Mar-2025,WBC8F81C195-1EE,Cash,Sale,Complete,$1.00,-179.720,$179.72,0,N/A

		execDate, err := time.Parse("02-Jan-2006", record[0])
		if err != nil {
			return nil, err
		}

		priceStr := strings.TrimPrefix(record[5], "$")
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}
		priceCents := int64(price * 100)

		quantity, err := strconv.ParseFloat(record[6], 64)
		if err != nil {
			return nil, err
		}
		// Quantity is negative in the CSV, so we make it positive
		if quantity < 0 {
			quantity = -quantity
		}

		sales = append(sales, models.Sale{
			Date:       execDate.Format("2006-01-02"),
			Quantity:   quantity,
			PriceCents: priceCents,
		})
	}

	return sales, nil
}
