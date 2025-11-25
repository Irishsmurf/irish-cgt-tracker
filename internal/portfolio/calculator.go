package portfolio

import (
	"fmt"
	"log"
	"math"

	"irish-cgt-tracker/internal/models"
)

// SettleSale calculates the chargeable gain or loss for a specific sale event
// according to Irish Capital Gains Tax rules, using the First-In, First-Out (FIFO)
// method for share matching.
//
// The process is as follows:
// 1. Fetches the details of the specified sale.
// 2. Retrieves the current inventory of vested shares that have not yet been fully sold.
// 3. Iterates through the vested shares, starting with the oldest (FIFO), matching them against the quantity of the current sale.
// 4. For each matched portion (a "lot"), it calculates the acquisition cost and disposal value in EUR, adhering to the "Irish Rule":
//    - Acquisition Cost in EUR = (USD Share Price at Vest * Quantity) * (EUR/USD rate on Vest Date)
//    - Disposal Value in EUR = (USD Share Price at Sale * Quantity) * (EUR/USD rate on Sale Date)
// 5. The created "lot" linking the sale and the vest is persisted to the database.
// 6. This continues until the entire sale quantity has been matched.
// 7. If the inventory is insufficient, it returns an error.
// 8. Finally, it marks the sale as "settled" and logs a summary of the calculation.
//
// Parameters:
//   - saleID: The unique identifier of the sale to be settled.
//
// Returns:
//   - An error if the sale is already settled, if there is a database issue, or if there is
//     insufficient vested inventory to cover the sale.
func (s *Service) SettleSale(saleID string) error {
	// 1. Fetch the sale details from the database.
	sale, err := s.getSale(saleID)
	if err != nil {
		return err
	}
	if sale.IsSettled {
		return fmt.Errorf("sale %s is already settled", saleID)
	}

	// 2. Get all vested shares, ordered by date, with their remaining unsold quantity.
	inventory, err := s.getAvailableInventory()
	if err != nil {
		return err
	}

	remainingToSell := sale.Quantity
	var totalCostBasisEUR float64
	var totalDisposalEUR float64

	log.Printf("--- Processing Sale: %d shares on %s ---", sale.Quantity, sale.Date)

	// 3. Match the sale quantity against the inventory using FIFO.
	for _, vest := range inventory {
		if remainingToSell <= 0 {
			break
		}
		if vest.RemainingQty <= 0 {
			continue // Skip fully used vests.
		}

		// Determine the number of shares to match from this vest.
		matchQty := int64(math.Min(float64(remainingToSell), float64(vest.RemainingQty)))

		// --- Apply Irish CGT Rules ---
		// Calculate cost basis for this lot in EUR.
		chunkCostUSD := float64(matchQty*vest.StrikePriceCents) / 100.0
		chunkCostEUR := chunkCostUSD * vest.ECBRate

		// Calculate disposal value for this lot in EUR.
		chunkSaleUSD := float64(matchQty*sale.PriceCents) / 100.0
		chunkSaleEUR := chunkSaleUSD * sale.ECBRate

		// Accumulate totals for the final report.
		totalCostBasisEUR += chunkCostEUR
		totalDisposalEUR += chunkSaleEUR

		// Persist the link between the sale and the vest (the "lot").
		if err := s.saveLot(sale.ID, vest.ID, matchQty); err != nil {
			return fmt.Errorf("failed to save lot: %w", err)
		}

		log.Printf("   Matched %d shares from Vest %s (Rate: %.4f)", matchQty, vest.Date, vest.ECBRate)
		remainingToSell -= matchQty
	}

	if remainingToSell > 0 {
		return fmt.Errorf("error: insufficient vest inventory to cover sale of %d shares (short by %d)", sale.Quantity, remainingToSell)
	}

	// 4. Mark the sale as fully processed.
	if err := s.markSaleSettled(sale.ID); err != nil {
		return err
	}

	// 5. Log the final calculation summary for this sale.
	gainLossEUR := totalDisposalEUR - totalCostBasisEUR
	log.Printf("=== RESULT for Sale %s ===", sale.Date)
	log.Printf("Disposal Value (EUR): €%.2f", totalDisposalEUR)
	log.Printf("Cost Basis (EUR):     €%.2f", totalCostBasisEUR)
	log.Printf("Chargeable Gain:      €%.2f", gainLossEUR)

	return nil
}

// InventoryItem is a helper struct that embeds a models.Vest and adds a field
// to track the remaining quantity of shares from that vest that have not yet
// been sold. This is used by the SettleSale calculator.
type InventoryItem struct {
	models.Vest
	RemainingQty int64
}
