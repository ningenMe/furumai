package kafka

import (
	"reflect"
	"testing"

	segmentio "github.com/segmentio/kafka-go"
)

func TestHeadersToMap(t *testing.T) {
	got := headersToMap([]segmentio.Header{
		{Key: "trace-id", Value: []byte("abc123")},
		{Key: "content-type", Value: []byte("application/json")},
	})

	want := map[string]string{
		"trace-id":     "abc123",
		"content-type": "application/json",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("headersToMap() = %v, want %v", got, want)
	}
}

func TestHeadersToMapEmpty(t *testing.T) {
	got := headersToMap(nil)
	if len(got) != 0 {
		t.Errorf("headersToMap(nil) = %v, want empty map", got)
	}
}
