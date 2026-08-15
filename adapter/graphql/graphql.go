// Package graphql is furumai's GraphQL Stimulus/Observation adapter.
//
// GraphQL over HTTP is just a JSON POST request and a {data, errors} JSON
// response, so this has no external dependency: net/http and
// encoding/json suffice.
package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Stimulus executes GraphQL operations against Endpoint. It is used from
// both given and when steps, since Stimulus adapters are shared between
// them.
type Stimulus struct {
	Endpoint string
	Client   *http.Client
}

// NewStimulus returns a Stimulus using http.DefaultClient.
func NewStimulus(endpoint string) *Stimulus {
	return &Stimulus{Endpoint: endpoint, Client: http.DefaultClient}
}

// RequestOption customizes a request before it is sent.
type RequestOption func(*http.Request)

// WithHeader adds a header to the request (e.g. Authorization).
func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) { r.Header.Add(key, value) }
}

// Response is the full-state Observation for a GraphQL operation. Data and
// Errors are typed any: their shape depends on the query/schema, and
// either can hold a furumai.Matcher (Any, Regex, ...) when building an
// expected Response.
type Response struct {
	Data   any
	Errors any
}

// Execute runs a query or mutation with variables and returns the
// full-state Response.
func (s *Stimulus) Execute(query string, variables map[string]any, opts ...RequestOption) (*Response, error) {
	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, s.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for _, opt := range opts {
		opt(req)
	}

	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed struct {
		Data   any `json:"data"`
		Errors any `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &Response{Data: parsed.Data, Errors: parsed.Errors}, nil
}
