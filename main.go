package main

import (
	"log"
	"os"

	"irish-cgt-tracker/internal/db"
	"irish-cgt-tracker/internal/portfolio"
)

func main() {
	// 1. Setup
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal(err)
	}
	// Clean start for this test run
	os.Remove("./data/portfolio.db") 
	
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()
	svc := portfolio.NewService(database)

	// 2. Setup Data (Matches your Product Spec Test Case)
	// Vest: 10 Units @ $100 on Jan 10 (Rate: ~0.93 on that day in 2023)
	// Note: In 2023, Jan 10 was Tue. 
	log.Println("--- Seed Data ---")
	vest, _ := svc.AddVest("2023-01-10", "TEST", 10, 10000) // $100.00
	
	// Sale: 10 Units @ $105 on June 10 (Sat -> Fri June 9 rate)
	sale, _ := svc.AddSale("2023-06-10", 10, 10500) // $105.00

	// 3. Run Calculation
	log.Println("--- Running FIFO Calculation ---")
	err := svc.SettleSale(sale.ID)
	if err != nil {
		log.Fatalf("Calculation failed: %v", err)
	}
	
	// Verification
	log.Printf("Used Vest Rate (Jan 10): %.4f", vest.ECBRate)
	log.Printf("Used Sale Rate (Jun 09): %.4f", sale.ECBRate)
}

