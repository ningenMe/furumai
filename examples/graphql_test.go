package examples

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/graphql"
)

// TestUserQuery demonstrates the GraphQL Stimulus/Observation adapter:
// when executes a query, then observes the full Response (data and
// errors) and compares it structurally against an expected value.
func TestUserQuery(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"user":{"id":1,"name":"Alice"}}}`))
	}))
	defer srv.Close()

	client := graphql.NewStimulus(srv.URL)

	var resp *graphql.Response

	furumai.When(t, func() error {
		var err error
		resp, err = client.Execute(`query($id: ID!) { user(id: $id) { id name } }`, map[string]any{"id": 1})
		return err
	})

	furumai.ThenEqual(t, *resp, graphql.Response{
		Data: map[string]any{
			"user": map[string]any{"id": float64(1), "name": "Alice"},
		},
		Errors: nil,
	})
}
