package db

import (
	"database/sql"
	"testing"
)

func TestInitDB(t *testing.T) {
	db, cleanup := NewTestDB(t)
	defer cleanup()

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
