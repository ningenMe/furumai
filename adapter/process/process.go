// Package process is furumai's shell command Stimulus/Observation adapter.
package process

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Stimulus runs shell commands. It is used from both given and when steps,
// since Stimulus adapters are shared between them.
type Stimulus struct{}

// NewStimulus returns a Stimulus.
func NewStimulus() *Stimulus { return &Stimulus{} }

// Result is the full-state Observation for a command run. Stdout and
// Stderr are typed any so a furumai.Matcher (Any, Regex, ...) can be
// substituted for either.
type Result struct {
	ExitCode int
	Stdout   any
	Stderr   any
}

type runOptions struct {
	env     []string
	dir     string
	stdin   string
	timeout time.Duration
}

// RunOption customizes a command run before it starts.
type RunOption func(*runOptions)

// WithEnv adds an environment variable, on top of the current process's
// environment.
func WithEnv(key, value string) RunOption {
	return func(o *runOptions) { o.env = append(o.env, key+"="+value) }
}

// WithDir sets the command's working directory.
func WithDir(dir string) RunOption {
	return func(o *runOptions) { o.dir = dir }
}

// WithStdin feeds input to the command's stdin.
func WithStdin(input string) RunOption {
	return func(o *runOptions) { o.stdin = input }
}

// WithTimeout kills the command if it hasn't finished after d.
func WithTimeout(d time.Duration) RunOption {
	return func(o *runOptions) { o.timeout = d }
}

// Run executes name with args and returns the full-state Result. A non-zero
// exit code is reported via Result.ExitCode, not a Go error: err is only
// non-nil when the command could not be run at all (not found, permission
// denied, ...).
func (s *Stimulus) Run(name string, args []string, opts ...RunOption) (*Result, error) {
	var o runOptions
	for _, opt := range opts {
		opt(&o)
	}

	ctx := context.Background()
	if o.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	if o.dir != "" {
		cmd.Dir = o.dir
	}
	if len(o.env) > 0 {
		cmd.Env = append(os.Environ(), o.env...)
	}
	if o.stdin != "" {
		cmd.Stdin = strings.NewReader(o.stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return &Result{
			ExitCode: exitErr.ExitCode(),
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &Result{
		ExitCode: 0,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}
