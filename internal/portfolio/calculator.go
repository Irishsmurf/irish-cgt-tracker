package portfolio

import (
	"fmt"
	"log"
	"math"

	"irish-cgt-tracker/internal/models"
)

// SettleSale calculates the gain/loss for a specific sale using FIFO.
func (s *Service) SettleSale(saleID string) error {
	// 1. Get the Sale details
	sale, err := s.getSale(saleID)
	if err != nil {
		return err
	}
	if sale.IsSettled {
		return fmt.Errorf("sale %s is already settled", saleID)
	}

	// 2. Get all Vests with their remaining (unsold) quantity
	inventory, err := s.getAvailableInventory()
	if err != nil {
		return err
	}

	remainingToSell := sale.Quantity
	var totalCostBasisEUR float64
	var totalDisposalEUR float64

	log.Printf("--- Processing Sale: %d shares on %s ---", sale.Quantity, sale.Date)

	// 3. FIFO Matching Loop
	for _, vest := range inventory {
		if remainingToSell <= 0 {
			break
		}

		if vest.RemainingQty <= 0 {
			continue // This vest is fully used up
		}

		// Take the smaller of: what we need to sell vs. what's in this vest
		matchQty := int64(math.Min(float64(remainingToSell), float64(vest.RemainingQty)))

		// --- THE IRISH MATH ---
		// 1. Cost Basis (EUR) for this specific chunk
		// (USD Price * Qty) * Vest Rate
		chunkCostUSD := float64(matchQty*vest.StrikePriceCents) / 100.0
		chunkCostEUR := chunkCostUSD * vest.ECBRate
		
		// 2. Disposal Value (EUR) for this specific chunk
		// (USD Price * Qty) * Sale Rate
		chunkSaleUSD := float64(matchQty*sale.PriceCents) / 100.0
		chunkSaleEUR := chunkSaleUSD * sale.ECBRate

		// Accumulate totals
		totalCostBasisEUR += chunkCostEUR
		totalDisposalEUR += chunkSaleEUR

		// Save the link (Lot) to DB
		err := s.saveLot(sale.ID, vest.ID, matchQty)
		if err != nil {
			return fmt.Errorf("failed to save lot: %w", err)
		}

		log.Printf("   Matched %d shares from Vest %s (Rate: %.4f)", matchQty, vest.Date, vest.ECBRate)
		
		remainingToSell -= matchQty
	}

	if remainingToSell > 0 {
		return fmt.Errorf("error: insufficient vest inventory to cover sale of %d shares (short by %d)", sale.Quantity, remainingToSell)
	}

	// 4. Mark Sale as Settled
	if err := s.markSaleSettled(sale.ID); err != nil {
		return err
	}

	// 5. Final Report for this Sale
	gainLossEUR := totalDisposalEUR - totalCostBasisEUR
	log.Printf("=== RESULT for Sale %s ===", sale.Date)
	log.Printf("Disposal Value (EUR): €%.2f", totalDisposalEUR)
	log.Printf("Cost Basis (EUR):     €%.2f", totalCostBasisEUR)
	log.Printf("Chargeable Gain:      €%.2f", gainLossEUR)
	
	return nil
}

// --- Helper Struct for Inventory ---
type InventoryItem struct {
	models.Vest
	RemainingQty int64
}
