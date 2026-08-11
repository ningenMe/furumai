package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ningenMe/furumai"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantExit   int
		wantStdout string
	}{
		{name: "version", args: []string{"version"}, wantExit: 0, wantStdout: "furumai " + version},
		{name: "help", args: []string{"help"}, wantExit: 0, wantStdout: "Usage: furumai <command>"},
		{name: "-h flag", args: []string{"-h"}, wantExit: 0, wantStdout: "Usage: furumai <command>"},
		{name: "no args", args: []string{}, wantExit: 1, wantStdout: "Usage: furumai <command>"},
		{name: "unknown command", args: []string{"nope"}, wantExit: 1, wantStdout: "Usage: furumai <command>"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			var gotExit int

			furumai.When(t, func() error {
				gotExit = run(&stdout, &stderr, tc.args)
				return nil
			})

			furumai.Then(t, func() error {
				if gotExit != tc.wantExit {
					return fmt.Errorf("exit code = %d, want %d", gotExit, tc.wantExit)
				}
				if !strings.Contains(stdout.String(), tc.wantStdout) {
					return fmt.Errorf("stdout = %q, want substring %q", stdout.String(), tc.wantStdout)
				}
				return nil
			})
		})
	}
}
