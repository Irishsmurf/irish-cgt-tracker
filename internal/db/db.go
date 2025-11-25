package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // Pure Go SQLite driver, CGO_ENABLED=0 friendly
)

// schema defines the SQL statements required to create the application's database structure.
// It includes tables for stock vests, sales, and a linking table to associate
// vested lots with sales in a many-to-many relationship, adhering to FIFO rules.
var schema = `
-- vests stores records of stock grants vesting.
-- ecb_rate is the USD to EUR exchange rate on the vesting date.
CREATE TABLE IF NOT EXISTS vests (
    id TEXT PRIMARY KEY,              -- Unique identifier for the vest
    date TEXT NOT NULL,               -- Vesting date (YYYY-MM-DD)
    symbol TEXT NOT NULL,             -- Stock ticker symbol (e.g., GOOGL)
    quantity INTEGER NOT NULL,        -- Number of shares vested
    strike_price_cents INTEGER NOT NULL, -- Price per share in USD cents at vest time
    ecb_rate REAL NOT NULL            -- USD to EUR ECB reference rate on the vest date
);

-- sales stores records of stock sales.
-- is_settled indicates if the CGT implications have been calculated and finalized.
CREATE TABLE IF NOT EXISTS sales (
    id TEXT PRIMARY KEY,              -- Unique identifier for the sale
    date TEXT NOT NULL,               -- Sale date (YYYY-MM-DD)
    quantity INTEGER NOT NULL,        -- Total number of shares sold
    price_cents INTEGER NOT NULL,     -- Price per share in USD cents at sale time
    ecb_rate REAL NOT NULL,           -- USD to EUR ECB reference rate on the sale date
    is_settled BOOLEAN NOT NULL DEFAULT 0 -- Flag for CGT calculation status
);

-- sale_lots links vests to sales, specifying how many shares from a
-- particular vest were included in a particular sale.
CREATE TABLE IF NOT EXISTS sale_lots (
    sale_id TEXT NOT NULL,            -- Foreign key to the sales table
    vest_id TEXT NOT NULL,            -- Foreign key to the vests table
    quantity INTEGER NOT NULL,        -- Number of shares from the vest lot used in this sale
    FOREIGN KEY(sale_id) REFERENCES sales(id),
    FOREIGN KEY(vest_id) REFERENCES vests(id),
    PRIMARY KEY (sale_id, vest_id)
);
`

// InitDB establishes a connection to a SQLite database at the given file path.
// If the database file does not exist, it will be created.
// It then ensures the necessary table schema is created by executing the statements
// in the global 'schema' variable.
// The function will terminate the application via log.Fatalf if the database
// connection or schema creation fails.
//
// Parameters:
//   - filepath: The path to the SQLite database file (e.g., "./data/portfolio.db").
//
// Returns:
//   - A pointer to an active sql.DB connection pool.
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Ensure the schema is created.
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database initialized successfully at", filepath)
	return db
}
