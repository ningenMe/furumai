package furumai

import (
	"testing"
	"time"
)

type row struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

// flexRow types CreatedAt as any so a Matcher can be substituted for it —
// a field can only ever hold a Matcher value if its static type allows it.
type flexRow struct {
	ID        int
	Name      string
	CreatedAt any
}

func TestDiffEqualStructs(t *testing.T) {
	got := row{ID: 1, Name: "Alice", CreatedAt: time.Unix(0, 0)}
	want := row{ID: 1, Name: "Alice", CreatedAt: time.Unix(0, 0)}

	if diffs := Diff(got, want); len(diffs) != 0 {
		t.Fatalf("Diff() = %v, want empty", diffs)
	}
}

func TestDiffReportsAllMismatches(t *testing.T) {
	got := row{ID: 1, Name: "Alice", CreatedAt: time.Unix(0, 0)}
	want := row{ID: 2, Name: "Bob", CreatedAt: time.Unix(0, 0)}

	diffs := Diff(got, want)
	if len(diffs) != 2 {
		t.Fatalf("Diff() = %v, want 2 mismatches (ID and Name)", diffs)
	}
}

func TestAnyMatchesAnything(t *testing.T) {
	got := flexRow{ID: 1, Name: "Alice", CreatedAt: time.Now()}
	want := flexRow{ID: 1, Name: "Alice", CreatedAt: Any()}

	if diffs := Diff(got, want); len(diffs) != 0 {
		t.Fatalf("Diff() = %v, want empty", diffs)
	}
}

func TestRegexMatcher(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want Matcher
		ok   bool
	}{
		{name: "match", got: "hello world", want: Regex(`^hello`), ok: true},
		{name: "no match", got: "goodbye", want: Regex(`^hello`), ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.want.Match(tc.got); got != tc.ok {
				t.Errorf("Match(%q) = %v, want %v", tc.got, got, tc.ok)
			}
		})
	}
}

func TestWithinMatcher(t *testing.T) {
	m := Within(10, 20)

	cases := []struct {
		name string
		got  any
		ok   bool
	}{
		{name: "in range", got: 15, ok: true},
		{name: "lower bound", got: 10, ok: true},
		{name: "upper bound", got: 20, ok: true},
		{name: "below range", got: 9, ok: false},
		{name: "above range", got: 21, ok: false},
		{name: "non-numeric", got: "15", ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := m.Match(tc.got); got != tc.ok {
				t.Errorf("Match(%v) = %v, want %v", tc.got, got, tc.ok)
			}
		})
	}
}

func TestDiffSliceOrdered(t *testing.T) {
	got := []int{1, 2, 3}
	want := []int{1, 3, 2}

	if diffs := Diff(got, want); len(diffs) == 0 {
		t.Fatal("Diff() = empty, want mismatches for out-of-order slice")
	}
}

func TestDiffSliceAnyOrder(t *testing.T) {
	got := []int{1, 2, 3}
	want := AnyOrder([]int{3, 1, 2})

	if diffs := Diff(got, want); len(diffs) != 0 {
		t.Fatalf("Diff() = %v, want empty", diffs)
	}
}

func TestDiffSliceAnyOrderMismatch(t *testing.T) {
	got := []int{1, 2, 3}
	want := AnyOrder([]int{1, 2, 4})

	if diffs := Diff(got, want); len(diffs) == 0 {
		t.Fatal("Diff() = empty, want mismatch (4 not present)")
	}
}

func TestDiffMapMissingAndUnexpectedKeys(t *testing.T) {
	got := map[string]int{"a": 1, "c": 3}
	want := map[string]int{"a": 1, "b": 2}

	diffs := Diff(got, want)
	if len(diffs) != 2 {
		t.Fatalf("Diff() = %v, want 2 mismatches (missing b, unexpected c)", diffs)
	}
}

func TestThenEqualReportsFailures(t *testing.T) {
	inner := &testing.T{}

	ThenEqual(inner, row{ID: 1, Name: "Alice"}, row{ID: 2, Name: "Alice"})

	if !inner.Failed() {
		t.Fatal("ThenEqual did not fail for mismatched values")
	}
}
