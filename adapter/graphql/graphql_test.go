package graphql

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestExecuteSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got, want := body.Variables["id"], float64(1); got != want {
			t.Fatalf("variables[id] = %v, want %v", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer abc"; got != want {
			t.Fatalf("Authorization header = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"user":{"id":1,"name":"Alice"}}}`))
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)

	resp, err := client.Execute(
		`query($id: ID!) { user(id: $id) { id name } }`,
		map[string]any{"id": 1},
		WithHeader("Authorization", "Bearer abc"),
	)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	want := map[string]any{
		"user": map[string]any{"id": float64(1), "name": "Alice"},
	}
	if !reflect.DeepEqual(resp.Data, want) {
		t.Errorf("Data = %#v, want %#v", resp.Data, want)
	}
	if resp.Errors != nil {
		t.Errorf("Errors = %v, want nil", resp.Errors)
	}
}

func TestExecuteGraphQLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":null,"errors":[{"message":"not found"}]}`))
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)

	resp, err := client.Execute(`query { missing }`, nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if resp.Data != nil {
		t.Errorf("Data = %v, want nil", resp.Data)
	}

	errs, ok := resp.Errors.([]any)
	if !ok || len(errs) != 1 {
		t.Fatalf("Errors = %#v, want a single-element slice", resp.Errors)
	}
}

func TestExecuteRequestBody(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Write([]byte(`{"data":{}}`))
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)
	if _, err := client.Execute("query { ping }", nil); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var sent struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("unmarshal sent body: %v", err)
	}
	if sent.Query != "query { ping }" {
		t.Errorf("sent query = %q, want %q", sent.Query, "query { ping }")
	}
}
