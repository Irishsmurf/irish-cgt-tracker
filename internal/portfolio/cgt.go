package portfolio

import (
	"fmt"
	"irish-cgt-tracker/internal/models"
	"log"
)

// SettleSale performs the FIFO logic to match a sale against the oldest available vests.
func (s *Service) SettleSale(saleID string) error {
	// 1. Fetch the sale details
	sale, err := s.getSale(saleID)
	if err != nil {
		return fmt.Errorf("could not retrieve sale %s: %w", saleID, err)
	}
	if sale.IsSettled {
		return fmt.Errorf("sale %s is already settled", saleID)
	}

	// 2. Fetch available inventory (vests with remaining shares), ordered by date (FIFO)
	inventory, err := s.getAvailableInventory()
	if err != nil {
		return fmt.Errorf("could not retrieve inventory: %w", err)
	}

	// 3. FIFO Logic
	sharesToSettle := sale.Quantity
	for _, vest := range inventory {
		if sharesToSettle == 0 {
			break
		}

		sharesToUse := min(sharesToSettle, vest.RemainingQty)
		if sharesToUse > 0 {
			// Create a "lot" linking this portion of the sale to this specific vest
			err := s.saveLot(sale.ID, vest.ID, sharesToUse)
			if err != nil {
				return fmt.Errorf("failed to save sale lot: %w", err)
			}

			// Perform the CGT calculation for this specific lot and save it
			err = s.calculateAndStoreCGT(sale, &vest.Vest, sharesToUse)
			if err != nil {
				return fmt.Errorf("failed to calculate CGT for lot: %w", err)
			}

			sharesToSettle -= sharesToUse
			log.Printf("Settled %d shares from sale %s against vest %s", sharesToUse, sale.ID, vest.ID)
		}
	}

	if sharesToSettle > 0 {
		return fmt.Errorf("insufficient shares available to settle sale %s. %d shares remain unsettled", sale.ID, sharesToSettle)
	}

	// 4. Mark the original sale as settled
	return s.markSaleSettled(sale.ID)
}

// calculateAndStoreCGT performs the core Irish CGT calculation for a single sale-vest lot.
func (s *Service) calculateAndStoreCGT(sale *models.Sale, vest *models.Vest, numShares int64) error {
	// All calculations are in cents to avoid floating point issues
	vestValuePerShare := float64(vest.StrikePriceCents)
	saleValuePerShare := float64(sale.PriceCents)

	// USD Calculations (per share)
	bookValueUSD := vestValuePerShare
	grossProceedUSD := saleValuePerShare
	gainLossUSD := grossProceedUSD - bookValueUSD

	// EUR Calculations (per share, applying the "Irish Rule")
	euroAcquisitionCost := bookValueUSD * vest.ECBRate
	euroDisposalValue := grossProceedUSD * sale.ECBRate
	euroGain := euroDisposalValue - euroAcquisitionCost

	// CGT @ 33%
	cgtTaxDue := euroGain * 0.33
	netProceeds := euroDisposalValue - cgtTaxDue

	// Create the record for the settled sale lot
	settledSale := models.SettledSale{
		SaleDate:           sale.Date,
		Ticker:             vest.Symbol,
		NumShares:          numShares,
		SalePriceUSD:       int64(saleValuePerShare * float64(numShares)),
		GainLossUSD:        int64(gainLossUSD * float64(numShares)),
		BookValueUSD:       int64(bookValueUSD * float64(numShares)),
		ExchangeRateAtVest: vest.ECBRate,
		GrossProceedUSD:    int64(grossProceedUSD * float64(numShares)),
		VestingValueUSD:    vest.StrikePriceCents * numShares,
		ExchangeRateAtSale: sale.ECBRate,
		EuroSaleEUR:        int64(euroDisposalValue * float64(numShares)),
		EuroGainEUR:        int64(euroGain * float64(numShares)),
		CGTTaxDueEUR:       int64(cgtTaxDue * float64(numShares)),
		Completed:          "Y",
		NetProceedsEUR:     int64(netProceeds * float64(numShares)),
		Type:               "FIFO",
	}

	// Persist to the new table
	return s.insertSettledSale(settledSale)
}

// insertSettledSale saves the calculated breakdown into the database.
func (s *Service) insertSettledSale(ss models.SettledSale) error {
	query := `
        INSERT INTO settled_sales (
            sale_date, ticker, num_shares, sale_price_usd, gain_loss_usd, book_value_usd,
            exchange_rate_at_vest, gross_proceed_usd, vesting_value_usd, exchange_rate_at_sale,
            euro_sale_eur, euro_gain_eur, cgt_tax_due_eur, completed, net_proceeds_eur, type
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query,
		ss.SaleDate, ss.Ticker, ss.NumShares, ss.SalePriceUSD, ss.GainLossUSD, ss.BookValueUSD,
		ss.ExchangeRateAtVest, ss.GrossProceedUSD, ss.VestingValueUSD, ss.ExchangeRateAtSale,
		ss.EuroSaleEUR, ss.EuroGainEUR, ss.CGTTaxDueEUR, ss.Completed, ss.NetProceedsEUR, ss.Type,
	)
	return err
}

// Simple min function for int64
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
