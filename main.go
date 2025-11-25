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

	// 3. Setup Web Server
	// We pass true for 'enableAuth' to turn on the protection we are about to build
	srv := server.NewServer(svc, true) 
	
	// Explicitly listen on 0.0.0.0 to allow external access (Docker/LAN)
	addr := "0.0.0.0:8080"
	log.Printf("Starting web server on %s...", addr)
	
	// Start
	srv.Start(addr)
}
