package examples

import (
	"os"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/postgres"
)

// TestUserSignupPostgres demonstrates the PostgreSQL Stimulus/Observation
// adapter. See examples/mysql_test.go for the MySQL equivalent; the two
// share the RDB capability shape.
//
// Requires a real PostgreSQL server; set POSTGRES_DSN to run it. CI
// provides this via a service container.
func TestUserSignupPostgres(t *testing.T) {
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_DSN not set; skipping PostgreSQL adapter example")
	}

	db, err := postgres.NewStimulus(dsn)
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

	got, err := db.Snapshot("users")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	furumai.ThenEqual(t, got, postgres.DataSet{
		"users": []postgres.Row{
			{"id": int32(1), "name": "Alice"},
		},
	})
}
