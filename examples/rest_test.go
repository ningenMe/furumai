package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/rest"
)

// TestGreetingAPI demonstrates the HTTP Stimulus/Observation adapter: when
// stimulates the system under test over HTTP, then observes the full
// Response and compares it structurally against an expected value.
func TestGreetingAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if name == "" {
			name = "world"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"greeting":"hello, ` + name + `"}`))
	}))
	defer srv.Close()

	client := rest.NewStimulus(srv.URL)

	var resp *rest.Response

	furumai.When(t, func() error {
		var err error
		resp, err = client.Get("/greeting", rest.WithQuery("name", "Alice"))
		return err
	})

	furumai.ThenEqual(t, *resp, rest.Response{
		StatusCode: http.StatusOK,
		Headers:    furumai.Ignore(),
		Body:       `{"greeting":"hello, Alice"}`,
	})
}
