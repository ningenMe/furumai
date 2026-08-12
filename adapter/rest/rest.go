// Package rest is furumai's HTTP Stimulus/Observation adapter.
package rest

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

// Stimulus sends HTTP requests against BaseURL. It is used from both given
// and when steps, since Stimulus adapters are shared between them.
type Stimulus struct {
	BaseURL string
	Client  *http.Client
}

// NewStimulus returns a Stimulus using http.DefaultClient.
func NewStimulus(baseURL string) *Stimulus {
	return &Stimulus{BaseURL: baseURL, Client: http.DefaultClient}
}

// Response is the full-state Observation for an HTTP request. Headers and
// Body are typed any so a furumai.Matcher (Any, Regex, ...) can be
// substituted for either when building an expected Response.
type Response struct {
	StatusCode int
	Headers    any
	Body       any
}

// RequestOption customizes a request before it is sent.
type RequestOption func(*http.Request)

// WithHeader adds a header to the request.
func WithHeader(key, value string) RequestOption {
	return func(r *http.Request) { r.Header.Add(key, value) }
}

// WithQuery adds a query parameter to the request.
func WithQuery(key, value string) RequestOption {
	return func(r *http.Request) {
		q := r.URL.Query()
		q.Add(key, value)
		r.URL.RawQuery = q.Encode()
	}
}

func (s *Stimulus) Get(path string, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodGet, path, nil, opts...)
}

func (s *Stimulus) Post(path string, body []byte, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodPost, path, body, opts...)
}

func (s *Stimulus) Put(path string, body []byte, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodPut, path, body, opts...)
}

func (s *Stimulus) Patch(path string, body []byte, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodPatch, path, body, opts...)
}

func (s *Stimulus) Delete(path string, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodDelete, path, nil, opts...)
}

func (s *Stimulus) Head(path string, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodHead, path, nil, opts...)
}

func (s *Stimulus) Options(path string, opts ...RequestOption) (*Response, error) {
	return s.do(http.MethodOptions, path, nil, opts...)
}

func (s *Stimulus) do(method, path string, body []byte, opts ...RequestOption) (*Response, error) {
	u, err := url.JoinPath(s.BaseURL, path)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, u, reader)
	if err != nil {
		return nil, err
	}
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(respBody),
	}, nil
}
