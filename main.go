package main

import (
	"log"
	"os"

	"irish-cgt-tracker/internal/db"
	"irish-cgt-tracker/internal/portfolio"
	"irish-cgt-tracker/internal/server"
)

// main is the entry point for the Irish CGT Tracker application.
//
// It performs the following steps:
// 1. Ensures the 'data' directory exists for the SQLite database.
// 2. Initializes the SQLite database connection and ensures its schema is up-to-date.
//    The database connection is deferred to close gracefully on application exit.
// 3. Creates an instance of the portfolio Service, which encapsulates the
//    core business logic of the application.
// 4. Creates and configures the web server, injecting the portfolio service
//    and enabling authentication.
// 5. Starts the web server, making it listen for incoming HTTP requests on
//    all network interfaces on port 8080.
func main() {
	// 1. Setup Database
	// Ensure the directory for the database file exists.
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	// Initialize the database connection and schema.
	database := db.InitDB("./data/portfolio.db")
	defer database.Close()

	// 2. Setup Core Application Logic
	// Instantiate the service layer with the database connection.
	svc := portfolio.NewService(database)

	// 3. Setup Web Server
	// Create a new server instance, enabling authentication.
	srv := server.NewServer(svc, true)

	// Define the server address. Listening on 0.0.0.0 makes it accessible
	// from outside its container or on the local network.
	addr := "0.0.0.0:8080"
	log.Printf("Starting web server on %s...", addr)

	// 4. Start Server
	// This function will block until the server is stopped.
	srv.Start(addr)
}
