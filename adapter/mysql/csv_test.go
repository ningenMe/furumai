package mysql

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadCSV(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "users.csv")
	content := "id,name,score\n1,Alice,9.5\n2,Bob,7\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := LoadCSV(path)
	if err != nil {
		t.Fatalf("LoadCSV: %v", err)
	}

	want := []Row{
		{"id": int64(1), "name": "Alice", "score": 9.5},
		{"id": int64(2), "name": "Bob", "score": int64(7)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadCSV() = %#v, want %#v", got, want)
	}
}

func TestInferType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want any
	}{
		{name: "integer", in: "42", want: int64(42)},
		{name: "float", in: "3.14", want: 3.14},
		{name: "string", in: "Alice", want: "Alice"},
		{name: "leading zero loses padding", in: "007", want: int64(7)},
		{name: "empty string stays string", in: "", want: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferType(tc.in); got != tc.want {
				t.Errorf("inferType(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}
