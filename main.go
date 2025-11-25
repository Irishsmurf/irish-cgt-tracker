package main

import (
	"log"
	"os"

	"irish-cgt-tracker/internal/currency"
	"irish-cgt-tracker/internal/db"
)

func main() {
	// 1. Initialize DB
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal(err)
	}
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()

	// 2. Test the Currency Service
	// Test Case A: A regular weekday (e.g., Friday, June 9th, 2023)
	rate, err := currency.FetchUSDToEUR("2023-06-09")
	if err != nil {
		log.Printf("Error fetching rate: %v", err)
	} else {
		log.Printf("Test A (Friday): Rate for 2023-06-09 is %.4f EUR", rate)
	}

	// Test Case B: A Sunday (e.g., June 11th, 2023) -> Should fallback to Friday June 9th
	rate, err = currency.FetchUSDToEUR("2023-06-11")
	if err != nil {
		log.Printf("Error fetching rate: %v", err)
	} else {
		log.Printf("Test B (Sunday): Rate for 2023-06-11 is %.4f EUR (Should match Friday)", rate)
	}

	log.Println("Server setup complete.")
}
