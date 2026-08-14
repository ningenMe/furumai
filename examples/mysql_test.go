package examples

import (
	"os"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/mysql"
)

// TestUserSignup demonstrates the MySQL Stimulus/Observation adapter:
// given/when seed and mutate table state, then queries the resulting rows
// and compares them structurally against an expected value.
//
// Requires a real MySQL server; set MYSQL_DSN to run it (see
// go-sql-driver/mysql's DSN format). CI provides this via a service
// container.
func TestUserSignup(t *testing.T) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		t.Skip("MYSQL_DSN not set; skipping MySQL adapter example")
	}

	db, err := mysql.NewStimulus(dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		name VARCHAR(255) NOT NULL
	)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	furumai.Given(t, func() error {
		return db.Truncate("users")
	})

	furumai.When(t, func() error {
		return db.Seed("users", map[string]any{"id": 1, "name": "Alice"})
	})

	got, err := db.Query("SELECT id, name FROM users ORDER BY id")
	if err != nil {
		t.Fatalf("query: %v", err)
	}

	furumai.ThenEqual(t, got, []mysql.Row{
		{"id": int64(1), "name": "Alice"},
	})
}
