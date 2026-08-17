package postgres

import "testing"

func TestNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   any
		want any
	}{
		{name: "bytes become string", in: []byte("Alice"), want: "Alice"},
		{name: "int64 passes through", in: int64(42), want: int64(42)},
		{name: "nil passes through", in: nil, want: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalize(tc.in); got != tc.want {
				t.Errorf("normalize(%#v) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}
