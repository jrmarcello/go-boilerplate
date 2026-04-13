package commands

import (
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout captures everything written to os.Stdout while fn runs.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("creating pipe: %v", pipeErr)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		b, readErr := io.ReadAll(r)
		if readErr != nil {
			done <- ""
			return
		}
		done <- string(b)
	}()

	fn()

	if closeErr := w.Close(); closeErr != nil {
		t.Fatalf("closing pipe writer: %v", closeErr)
	}
	os.Stdout = original
	return <-done
}

func TestCheckTool_GoInstalled(t *testing.T) {
	out := captureStdout(t, func() {
		checkTool(toolCheck{
			name:        "Go",
			cmd:         "go",
			versionArgs: []string{"version"},
			installHint: "https://go.dev/doc/install",
		})
	})

	if !strings.Contains(out, "Go") {
		t.Errorf("expected output to mention Go, got: %q", out)
	}
	if !strings.Contains(out, "[OK]") {
		t.Errorf("expected [OK] marker for installed Go, got: %q", out)
	}
}

func TestCheckTool_NonExistent(t *testing.T) {
	out := captureStdout(t, func() {
		checkTool(toolCheck{
			name:        "xyzabc",
			cmd:         "xyzabc-does-not-exist",
			versionArgs: []string{"--version"},
			installHint: "brew install xyzabc",
		})
	})

	if !strings.Contains(out, "[!!]") {
		t.Errorf("expected [!!] marker for missing tool, got: %q", out)
	}
	if !strings.Contains(out, "xyzabc") {
		t.Errorf("expected tool name in output, got: %q", out)
	}
	if !strings.Contains(out, "brew install xyzabc") {
		t.Errorf("expected install hint in output, got: %q", out)
	}
}

func TestCheckProject_NoGoMod(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)

	out := captureStdout(t, func() {
		checkProject()
	})

	if !strings.Contains(out, "[!!]") {
		t.Errorf("expected [!!] marker when go.mod is missing, got: %q", out)
	}
	if !strings.Contains(out, "go.mod not found") {
		t.Errorf("expected 'go.mod not found' message, got: %q", out)
	}
}

func TestCheckContainer_Running(t *testing.T) {
	out := captureStdout(t, func() {
		checkContainer("my-postgres-db\nmy-redis-cache", "postgres")
	})

	if !strings.Contains(out, "[OK]") {
		t.Errorf("expected [OK] for running container, got: %q", out)
	}
	if !strings.Contains(out, "postgres running") {
		t.Errorf("expected 'postgres running' in output, got: %q", out)
	}
}

func TestCheckContainer_NotRunning(t *testing.T) {
	out := captureStdout(t, func() {
		checkContainer("other-container", "postgres")
	})

	if !strings.Contains(out, "[--]") {
		t.Errorf("expected [--] for missing container, got: %q", out)
	}
	if !strings.Contains(out, "postgres not running") {
		t.Errorf("expected 'postgres not running' in output, got: %q", out)
	}
}

func TestRunDoctor_NoError(t *testing.T) {
	out := captureStdout(t, func() {
		if runErr := runDoctor(nil, nil); runErr != nil {
			t.Errorf("runDoctor returned error: %v", runErr)
		}
	})

	if !strings.Contains(out, "gopherplate doctor") {
		t.Errorf("expected header 'gopherplate doctor', got: %q", out)
	}
	if !strings.Contains(out, "Project:") {
		t.Errorf("expected 'Project:' section, got: %q", out)
	}
}
