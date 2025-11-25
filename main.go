package main

import (
	"log"
	"os"

	"irish-cgt-tracker/internal/db"
)

func main() {
	// Ensure the data directory exists (simulating the volume mount)
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal(err)
	}

	// Initialize Database
	// In production, this path will be inside the container volume
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()

	log.Println("Server is ready to start...")
	
	// Future: Start HTTP server here
}
