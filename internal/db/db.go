package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

var schema = `
CREATE TABLE IF NOT EXISTS vests (
    id TEXT PRIMARY KEY,
    date TEXT NOT NULL,
    symbol TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    strike_price_cents INTEGER NOT NULL,
    ecb_rate REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS sales (
    id TEXT PRIMARY KEY,
    date TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price_cents INTEGER NOT NULL,
    ecb_rate REAL NOT NULL,
    is_settled BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS sale_lots (
    sale_id TEXT NOT NULL,
    vest_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    FOREIGN KEY(sale_id) REFERENCES sales(id),
    FOREIGN KEY(vest_id) REFERENCES vests(id),
    PRIMARY KEY (sale_id, vest_id)
);

CREATE TABLE IF NOT EXISTS settled_sales (
    sale_date TEXT,
    ticker TEXT,
    num_shares INTEGER,
    sale_price_usd INTEGER,
    gain_loss_usd INTEGER,
    book_value_usd INTEGER,
    exchange_rate_at_vest REAL,
    gross_proceed_usd INTEGER,
    vesting_value_usd INTEGER,
    exchange_rate_at_sale REAL,
    euro_sale_eur INTEGER,
    euro_gain_eur INTEGER,
    cgt_tax_due_eur INTEGER,
    completed TEXT,
    net_proceeds_eur INTEGER,
    type TEXT
);
`

// InitDB creates the database file and runs migrations
func InitDB(filepath string) *sql.DB {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// create tables
	if _, err := db.Exec(schema); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	log.Println("Database initialized successfully at", filepath)
	return db
}
