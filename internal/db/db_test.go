package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestInitDB(t *testing.T) {
	// Use an in-memory SQLite database for testing
	db := InitDB(":memory:")
	if db == nil {
		t.Fatal("expected a database connection, but got nil")
	}
	defer db.Close()

	// Check if tables were created
	tables := []string{"vests", "sales", "sale_lots"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			if err == sql.ErrNoRows {
				t.Errorf("table %s was not created", table)
			} else {
				t.Errorf("error checking for table %s: %s", table, err)
			}
		}
	}
}

func TestInitDB_File(t *testing.T) {
	// Use a temporary file for the database
	filepath := "test.db"
	db := InitDB(filepath)
	if db == nil {
		t.Fatal("expected a database connection, but got nil")
	}
	defer db.Close()
	defer os.Remove(filepath)

	// Check if the file was created
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("database file %s was not created", filepath)
	}
}
