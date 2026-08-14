package examples

import (
	"testing"

	"github.com/ningenMe/furumai"
	"github.com/ningenMe/furumai/adapter/process"
)

// TestGreetingScript demonstrates the shell command Stimulus/Observation
// adapter: when runs a command, then observes the full Result (exit code,
// stdout, stderr) and compares it structurally against an expected value.
func TestGreetingScript(t *testing.T) {
	proc := process.NewStimulus()

	var result *process.Result

	furumai.When(t, func() error {
		var err error
		result, err = proc.Run("sh", []string{"-c", "echo hello, Alice"})
		return err
	})

	furumai.ThenEqual(t, *result, process.Result{
		ExitCode: 0,
		Stdout:   "hello, Alice\n",
		Stderr:   "",
	})
}
