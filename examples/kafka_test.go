package examples

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/kafka"
)

// TestOrderPlaced demonstrates the Kafka Stimulus/Observation adapter:
// Listen starts observing a topic before when publishes to it, then
// Collect gathers what arrived within a window and compares it
// structurally against an expected value.
//
// Requires a real Kafka broker; set KAFKA_BROKERS (comma-separated) to run
// it. CI provides this via a service container.
func TestOrderPlaced(t *testing.T) {
	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		t.Skip("KAFKA_BROKERS not set; skipping Kafka adapter example")
	}
	brokers := strings.Split(brokersEnv, ",")

	mq := kafka.NewStimulus(brokers...)

	sub, err := mq.Listen("orders")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer sub.Close()

	furumai.When(t, func() error {
		return mq.PublishJSON("orders", "order-1", map[string]any{"id": 1, "item": "widget"})
	})

	got, err := sub.Collect(5 * time.Second)
	if err != nil {
		t.Fatalf("Collect: %v", err)
	}

	furumai.ThenEqual(t, got, []kafka.Message{
		{
			Key:     "order-1",
			Value:   `{"id":1,"item":"widget"}`,
			Headers: furumai.Ignore(),
		},
	})
}
