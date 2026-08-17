package examples

import (
	"os"
	"strings"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/cassandra"
)

// TestUserSignupCassandra demonstrates the Cassandra Stimulus/Observation
// adapter: given/when write row state via CQL, then Snapshot fetches the
// resulting full table state and compares it structurally against an
// expected value.
//
// Requires a real Cassandra cluster; set CASSANDRA_HOSTS (comma-separated)
// to run it. CI provides this via a service container. The keyspace must
// already exist (furumai_test, replication factor 1); this example creates
// it if missing.
func TestUserSignupCassandra(t *testing.T) {
	hostsEnv := os.Getenv("CASSANDRA_HOSTS")
	if hostsEnv == "" {
		t.Skip("CASSANDRA_HOSTS not set; skipping Cassandra adapter example")
	}
	hosts := strings.Split(hostsEnv, ",")

	admin, err := cassandra.NewStimulus(hosts...)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer admin.Close()

	if err := admin.Exec(`CREATE KEYSPACE IF NOT EXISTS furumai_test
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}`); err != nil {
		t.Fatalf("create keyspace: %v", err)
	}
	if err := admin.Exec(`CREATE TABLE IF NOT EXISTS furumai_test.users (
		id int PRIMARY KEY,
		name text
	)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	db := admin

	furumai.Given(t, func() error {
		return db.Exec("TRUNCATE furumai_test.users")
	})

	furumai.When(t, func() error {
		return db.Exec("INSERT INTO furumai_test.users (id, name) VALUES (?, ?)", 1, "Alice")
	})

	got, err := db.Snapshot(cassandra.TableQuery{Table: "furumai_test.users"})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	furumai.ThenEqual(t, got, cassandra.DataSet{
		"furumai_test.users": []cassandra.Row{
			{"id": 1, "name": "Alice"},
		},
	})
}
