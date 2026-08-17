package examples

import (
	"os"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/mysql"
)

// TestUserSignup demonstrates the MySQL Stimulus/Observation adapter:
// given/when seed and mutate table state, then Snapshot fetches the
// resulting full table state and compares it structurally against an
// expected value.
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

	fixture, err := mysql.LoadCSV("testdata/mysql_users.csv")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	seedRows := make([]map[string]any, len(fixture))
	for i, row := range fixture {
		seedRows[i] = row
	}

	furumai.Given(t, func() error {
		return db.Truncate("users")
	})

	furumai.When(t, func() error {
		return db.Seed("users", seedRows...)
	})

	got, err := db.Snapshot("users")
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	// MySQL doesn't guarantee row order without an ORDER BY, so the
	// expected rows (loaded from the same static fixture used to seed)
	// are compared as a multiset via AnyOrder rather than by position.
	furumai.ThenEqual(t, got, mysql.DataSet{
		"users": furumai.AnyOrder(fixture),
	})
}
