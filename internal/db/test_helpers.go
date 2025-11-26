package db

import (
	"database/sql"
	"testing"
)

// NewTestDB creates a new in-memory SQLite database for testing.
// It returns the database connection and a cleanup function that should be called
// when the test is complete.
func NewTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db := InitDB(":memory:")
	if db == nil {
		t.Fatal("expected a database connection, but got nil")
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}
