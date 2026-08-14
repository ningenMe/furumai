package process

import (
	"strings"
	"testing"
	"time"
)

func TestRunSuccess(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("sh", []string{"-c", "echo hello"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "hello\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "hello\n")
	}
}

func TestRunNonZeroExit(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("sh", []string{"-c", "echo oops >&2; exit 3"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", result.ExitCode)
	}
	if result.Stderr != "oops\n" {
		t.Errorf("Stderr = %q, want %q", result.Stderr, "oops\n")
	}
}

func TestRunCommandNotFound(t *testing.T) {
	s := NewStimulus()

	if _, err := s.Run("furumai-does-not-exist", nil); err == nil {
		t.Fatal("Run() = nil error, want an error for a missing command")
	}
}

func TestRunWithEnv(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("sh", []string{"-c", "echo $FOO"}, WithEnv("FOO", "bar"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Stdout != "bar\n" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "bar\n")
	}
}

func TestRunWithStdin(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("cat", nil, WithStdin("piped input"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.Stdout != "piped input" {
		t.Errorf("Stdout = %q, want %q", result.Stdout, "piped input")
	}
}

func TestRunWithDir(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("pwd", nil, WithDir("/tmp"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	got := strings.TrimSpace(result.Stdout.(string))
	if got != "/tmp" {
		t.Errorf("pwd = %q, want %q", got, "/tmp")
	}
}

func TestRunWithTimeout(t *testing.T) {
	s := NewStimulus()

	result, err := s.Run("sleep", []string{"5"}, WithTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if result.ExitCode == 0 {
		t.Errorf("ExitCode = 0, want non-zero (command should have been killed by timeout)")
	}
}
