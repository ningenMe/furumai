package furumai

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

// Matcher matches a single value during a structural diff, in place of a
// literal expected value. A struct field can only ever hold a Matcher if
// its static type allows it (typically declared as any), since Go has no
// way to substitute a different type into a concretely-typed field.
type Matcher interface {
	Match(got any) bool
	String() string
}

type anyMatcher struct{}

func (anyMatcher) Match(any) bool { return true }
func (anyMatcher) String() string { return "Any()" }

// Any matches any value; it only asserts presence.
func Any() Matcher { return anyMatcher{} }

// Ignore excludes a value from comparison. Distinct from Any for
// readability at call sites.
func Ignore() Matcher { return anyMatcher{} }

type regexMatcher struct {
	re *regexp.Regexp
}

func (m regexMatcher) Match(got any) bool {
	s, ok := got.(string)
	if !ok {
		return false
	}
	return m.re.MatchString(s)
}

func (m regexMatcher) String() string { return fmt.Sprintf("Regex(%q)", m.re.String()) }

// Regex matches string values against pattern.
func Regex(pattern string) Matcher {
	return regexMatcher{re: regexp.MustCompile(pattern)}
}

type withinMatcher struct {
	min, max float64
}

func (m withinMatcher) Match(got any) bool {
	v, ok := toFloat64(got)
	if !ok {
		return false
	}
	return v >= m.min && v <= m.max
}

func (m withinMatcher) String() string {
	return fmt.Sprintf("Within(%v, %v)", m.min, m.max)
}

// Within matches numeric values in the inclusive range [min, max].
func Within(min, max float64) Matcher {
	return withinMatcher{min: min, max: max}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// anyOrderWant wraps a slice expectation so Diff compares it as a multiset
// instead of comparing elements by index.
type anyOrderWant struct {
	want any
}

// AnyOrder marks a slice expectation as order-independent.
func AnyOrder(want any) any {
	return anyOrderWant{want: want}
}

// Diff compares got against want and returns a description of every
// mismatch found, in no particular order. An empty result means got
// matches want.
//
// want may contain Matcher values (see Any, Regex, Within, Ignore) in
// place of literal expected values, and slices wrapped with AnyOrder are
// compared without regard to element order.
func Diff(got, want any) []string {
	return diffValue("$", reflect.ValueOf(got), reflect.ValueOf(want))
}

func diffValue(path string, got, want reflect.Value) []string {
	if want.IsValid() && want.CanInterface() {
		if aow, ok := want.Interface().(anyOrderWant); ok {
			return diffAnyOrder(path, got, reflect.ValueOf(aow.want))
		}
		if m, ok := want.Interface().(Matcher); ok {
			var gotIface any
			if got.IsValid() {
				gotIface = got.Interface()
			}
			if !m.Match(gotIface) {
				return []string{fmt.Sprintf("%s: got %v, want match %s", path, gotIface, m)}
			}
			return nil
		}
	}

	if !got.IsValid() && !want.IsValid() {
		return nil
	}
	if !got.IsValid() || !want.IsValid() {
		return []string{fmt.Sprintf("%s: got %s, want %s", path, describe(got), describe(want))}
	}

	got = deref(got)
	want = deref(want)

	if got.Type() != want.Type() {
		return []string{fmt.Sprintf("%s: type mismatch: got %s, want %s", path, got.Type(), want.Type())}
	}

	switch want.Kind() {
	case reflect.Struct:
		var diffs []string
		for i := 0; i < want.NumField(); i++ {
			field := want.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			diffs = append(diffs, diffValue(path+"."+field.Name, got.Field(i), want.Field(i))...)
		}
		return diffs
	case reflect.Slice, reflect.Array:
		var diffs []string
		if got.Len() != want.Len() {
			diffs = append(diffs, fmt.Sprintf("%s: length mismatch: got %d, want %d", path, got.Len(), want.Len()))
		}
		n := min(got.Len(), want.Len())
		for i := 0; i < n; i++ {
			diffs = append(diffs, diffValue(fmt.Sprintf("%s[%d]", path, i), got.Index(i), want.Index(i))...)
		}
		return diffs
	case reflect.Map:
		var diffs []string
		for _, key := range want.MapKeys() {
			gotVal := got.MapIndex(key)
			if !gotVal.IsValid() {
				diffs = append(diffs, fmt.Sprintf("%s[%v]: missing key", path, key.Interface()))
				continue
			}
			diffs = append(diffs, diffValue(fmt.Sprintf("%s[%v]", path, key.Interface()), gotVal, want.MapIndex(key))...)
		}
		for _, key := range got.MapKeys() {
			if !want.MapIndex(key).IsValid() {
				diffs = append(diffs, fmt.Sprintf("%s[%v]: unexpected key", path, key.Interface()))
			}
		}
		return diffs
	default:
		if !reflect.DeepEqual(got.Interface(), want.Interface()) {
			return []string{fmt.Sprintf("%s: got %v, want %v", path, got.Interface(), want.Interface())}
		}
		return nil
	}
}

// diffAnyOrder compares got against want as multisets: every element of
// want must match exactly one distinct element of got, regardless of
// position.
func diffAnyOrder(path string, got, want reflect.Value) []string {
	got = deref(got)
	want = deref(want)

	if got.Kind() != reflect.Slice && got.Kind() != reflect.Array {
		return []string{fmt.Sprintf("%s: got %s, want a slice (AnyOrder)", path, describe(got))}
	}
	if want.Kind() != reflect.Slice && want.Kind() != reflect.Array {
		return []string{fmt.Sprintf("%s: AnyOrder must wrap a slice, got %s", path, describe(want))}
	}

	used := make([]bool, got.Len())
	var diffs []string
	for i := 0; i < want.Len(); i++ {
		found := false
		for j := 0; j < got.Len(); j++ {
			if used[j] {
				continue
			}
			if len(diffValue("$", got.Index(j), want.Index(i))) == 0 {
				used[j] = true
				found = true
				break
			}
		}
		if !found {
			diffs = append(diffs, fmt.Sprintf("%s: no element matches %v", path, want.Index(i).Interface()))
		}
	}
	if got.Len() != want.Len() {
		diffs = append(diffs, fmt.Sprintf("%s: length mismatch: got %d, want %d", path, got.Len(), want.Len()))
	}
	return diffs
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

func describe(v reflect.Value) string {
	if !v.IsValid() {
		return "<invalid>"
	}
	return fmt.Sprintf("%v", v.Interface())
}

// ThenEqual observes got and reports a failure listing every mismatch
// against want (see Diff).
func ThenEqual(t *testing.T, got, want any) {
	t.Helper()
	if diffs := Diff(got, want); len(diffs) > 0 {
		for _, d := range diffs {
			t.Errorf("then: %s", d)
		}
	}
}
