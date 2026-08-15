package examples

import (
	"context"
	"os"
	"testing"

	awsdynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/dynamodb"
)

// TestUserSignupDynamoDB demonstrates the DynamoDB Stimulus/Observation
// adapter: when puts an item, then GetItem observes the full Item and
// compares it structurally against an expected value.
//
// Requires a DynamoDB-compatible endpoint; set DYNAMODB_ENDPOINT (e.g.
// http://localhost:4566 for LocalStack) to run it. CI provides this via a
// LocalStack service container, with a "users" table (partition key "id",
// type N) pre-created.
func TestUserSignupDynamoDB(t *testing.T) {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		t.Skip("DYNAMODB_ENDPOINT not set; skipping DynamoDB adapter example")
	}

	ctx := context.Background()
	db, err := dynamodb.NewStimulus(ctx, func(o *awsdynamodb.Options) {
		o.BaseEndpoint = &endpoint
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	furumai.Given(t, func() error {
		return db.DeleteItem(ctx, "users", map[string]any{"id": 1})
	})

	furumai.When(t, func() error {
		return db.PutItem(ctx, "users", map[string]any{"id": 1, "name": "Alice"})
	})

	got, err := db.GetItem(ctx, "users", map[string]any{"id": 1})
	if err != nil {
		t.Fatalf("GetItem: %v", err)
	}

	furumai.ThenEqual(t, got, dynamodb.Item{
		"id":   float64(1),
		"name": "Alice",
	})
}
