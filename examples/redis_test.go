package examples

import (
	"os"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/redis"
)

// TestSessionCache demonstrates the Redis Stimulus/Observation adapter:
// given/when write key state, then Get observes the full set of keys
// matching a pattern and compares it structurally against an expected
// value.
//
// Requires a real Redis server; set REDIS_ADDR (host:port) to run it. CI
// provides this via a service container.
func TestSessionCache(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR not set; skipping Redis adapter example")
	}

	kv, err := redis.NewStimulus(addr)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer kv.Close()

	furumai.Given(t, func() error {
		return kv.Del("session:1")
	})

	furumai.When(t, func() error {
		return kv.Set("session:1", "alice")
	})

	got, err := kv.Get("session:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	furumai.ThenEqual(t, got, map[string]redis.Value{
		"session:1": "alice",
	})
}
