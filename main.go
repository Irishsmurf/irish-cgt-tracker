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
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()

	svc := portfolio.NewService(database)

	// 2. Simulate User Input: Adding a Vest (Acquisition)
	// Example: 100 shares of GOOG vest on a specific date
	log.Println("--- Adding Vest ---")
	_, err := svc.AddVest("2023-01-15", "GOOG", 100, 9500) // $95.00
	if err != nil {
		log.Printf("Error adding vest: %v", err)
	}

	// 3. Simulate User Input: Adding a Sale (Disposal)
	// Example: Selling 50 shares later in the year
	log.Println("--- Adding Sale ---")
	_, err = svc.AddSale("2023-06-10", 50, 12000) // $120.00
	if err != nil {
		log.Printf("Error adding sale: %v", err)
	}

	log.Println("--- Phase 1 Complete: Data persisted with Irish-compliant rates ---")
}
