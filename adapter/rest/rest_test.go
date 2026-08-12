package rest

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ningenMe/furumai"
)

func TestStimulusGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/greeting" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got, want := r.URL.Query().Get("name"), "Alice"; got != want {
			t.Fatalf("query param name = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("X-Test"), "abc"; got != want {
			t.Fatalf("header X-Test = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"greeting":"hello, Alice"}`))
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)

	var resp *Response
	furumai.Given(t, func() error { return nil })

	furumai.When(t, func() error {
		var err error
		resp, err = client.Get("/greeting",
			WithQuery("name", "Alice"),
			WithHeader("X-Test", "abc"),
		)
		return err
	})

	furumai.ThenEqual(t, *resp, Response{
		StatusCode: http.StatusOK,
		Headers:    furumai.Any(),
		Body:       `{"greeting":"hello, Alice"}`,
	})
}

func TestStimulusPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/users" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if got, want := string(body), `{"name":"Bob"}`; got != want {
			t.Fatalf("request body = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1,"name":"Bob"}`))
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)

	var resp *Response
	furumai.When(t, func() error {
		var err error
		resp, err = client.Post("/users", []byte(`{"name":"Bob"}`))
		return err
	})

	furumai.ThenEqual(t, *resp, Response{
		StatusCode: http.StatusCreated,
		Headers:    furumai.Ignore(),
		Body:       `{"id":1,"name":"Bob"}`,
	})
}

func TestStimulusMismatchReportsDiff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewStimulus(srv.URL)

	resp, err := client.Get("/missing")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	diffs := furumai.Diff(*resp, Response{StatusCode: http.StatusOK, Headers: furumai.Ignore(), Body: furumai.Ignore()})
	if len(diffs) == 0 {
		t.Fatal("Diff() = empty, want a StatusCode mismatch")
	}
}
