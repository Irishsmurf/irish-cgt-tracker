package main

import (
	"log"
	"os"

	"irish-cgt-tracker/internal/db"
	"irish-cgt-tracker/internal/portfolio"
	"irish-cgt-tracker/internal/server"
)

func main() {
	// 1. Setup DB
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal(err)
	}
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()

	// 2. Setup Logic
	svc := portfolio.NewService(database)

	// 3. Setup and Start Web Server
	srv := server.NewServer(svc)
	
	log.Println("Starting web server...")
	log.Println("Go to http://coventry.paddez.com:8020")
	srv.Start(":8020")
}
