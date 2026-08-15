package examples

import (
	"context"
	"os"
	"testing"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/s3"
)

// TestUploadReceipt demonstrates the S3 Stimulus/Observation adapter: when
// puts an object, then GetObject observes the full Object (content,
// metadata) and compares it structurally against an expected value.
//
// Requires an S3-compatible endpoint; set S3_ENDPOINT (e.g.
// http://localhost:4566 for LocalStack) and S3_BUCKET to run it. CI
// provides this via a LocalStack service container.
func TestUploadReceipt(t *testing.T) {
	endpoint := os.Getenv("S3_ENDPOINT")
	bucket := os.Getenv("S3_BUCKET")
	if endpoint == "" || bucket == "" {
		t.Skip("S3_ENDPOINT/S3_BUCKET not set; skipping S3 adapter example")
	}

	ctx := context.Background()
	store, err := s3.NewStimulus(ctx, bucket, func(o *awss3.Options) {
		o.BaseEndpoint = &endpoint
		o.UsePathStyle = true
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	furumai.When(t, func() error {
		return store.PutObject(ctx, "receipts/1.txt", []byte("thanks for your order"))
	})

	got, err := store.GetObject(ctx, "receipts/1.txt")
	if err != nil {
		t.Fatalf("GetObject: %v", err)
	}

	furumai.ThenEqual(t, *got, s3.Object{
		Key:      "receipts/1.txt",
		Content:  "thanks for your order",
		Metadata: furumai.Ignore(),
	})
}
