// Package furumai provides the given/when/then test-definition DSL.
package furumai

import "testing"

// Given runs a setup step. A failing step aborts the test immediately,
// since later steps generally depend on it having succeeded.
func Given(t *testing.T, fn func() error) {
	t.Helper()
	if err := fn(); err != nil {
		t.Fatalf("given: %v", err)
	}
}

// When stimulates the system under test. A failing step aborts the test
// immediately, since Then steps have nothing to observe otherwise.
func When(t *testing.T, fn func() error) {
	t.Helper()
	if err := fn(); err != nil {
		t.Fatalf("when: %v", err)
	}
}

// Then observes and verifies the result. A failing step marks the test as
// failed but lets remaining Then steps still run, so all assertions for a
// scenario get reported together.
func Then(t *testing.T, fn func() error) {
	t.Helper()
	if err := fn(); err != nil {
		t.Errorf("then: %v", err)
	}
}
