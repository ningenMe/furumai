package redis

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestEncodeCommand(t *testing.T) {
	got := string(encodeCommand([]string{"SET", "foo", "bar"}))
	want := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	if got != want {
		t.Errorf("encodeCommand() = %q, want %q", got, want)
	}
}

func TestReadReply(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want any
	}{
		{name: "simple string", in: "+OK\r\n", want: "OK"},
		{name: "integer", in: ":1000\r\n", want: int64(1000)},
		{name: "bulk string", in: "$6\r\nfoobar\r\n", want: "foobar"},
		{name: "empty bulk string", in: "$0\r\n\r\n", want: ""},
		{name: "nil bulk string", in: "$-1\r\n", want: nil},
		{name: "nil array", in: "*-1\r\n", want: nil},
		{name: "empty array", in: "*0\r\n", want: []any{}},
		{
			name: "array of bulk strings",
			in:   "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			want: []any{"hello", "world"},
		},
		{
			name: "nested array",
			in:   "*2\r\n*2\r\n:1\r\n:2\r\n$3\r\nfoo\r\n",
			want: []any{[]any{int64(1), int64(2)}, "foo"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := readReply(bufio.NewReader(strings.NewReader(tc.in)))
			if err != nil {
				t.Fatalf("readReply(%q): %v", tc.in, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("readReply(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}

func TestReadReplyError(t *testing.T) {
	_, err := readReply(bufio.NewReader(strings.NewReader("-ERR unknown command\r\n")))
	if err == nil {
		t.Fatal("readReply() = nil error, want an error for a RESP error reply")
	}
}

func TestToStringMap(t *testing.T) {
	got := toStringMap([]any{"name", "Alice", "age", "30"})
	want := map[string]string{"name": "Alice", "age": "30"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("toStringMap() = %v, want %v", got, want)
	}
}

func TestToScoreMap(t *testing.T) {
	got := toScoreMap([]any{"alice", "1.5", "bob", "2"})
	want := map[string]float64{"alice": 1.5, "bob": 2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("toScoreMap() = %v, want %v", got, want)
	}
}

func TestFlushDBRequiresConfirm(t *testing.T) {
	s := &Stimulus{}
	if err := s.FlushDB(false); err == nil {
		t.Fatal("FlushDB(false) = nil error, want an error")
	}
}
