// Package examples demonstrates the given/when/then DSL. Stimulus and
// observation here are plain function calls rather than real adapters
// (HTTP, DB, Kafka, ...), which are added in later tasks.
package examples

import (
	"fmt"
	"testing"

	"github.com/ningenMe/furumai"
)

func greet(name string) string {
	if name == "" {
		return "hello, world"
	}
	return "hello, " + name
}

func TestGreeting(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "known name", input: "Alice", want: "hello, Alice"},
		{name: "empty name", input: "", want: "hello, world"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string

			furumai.Given(t, func() error {
				got = ""
				return nil
			})

			furumai.When(t, func() error {
				got = greet(tc.input)
				return nil
			})

			furumai.Then(t, func() error {
				if got != tc.want {
					return fmt.Errorf("got %q, want %q", got, tc.want)
				}
				return nil
			})
		})
	}
}
