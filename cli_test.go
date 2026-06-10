package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestDispatchCLIVersion(t *testing.T) {
	exit, code := dispatchCLI([]string{"docklog", "version"})
	if !exit || code != 0 {
		t.Fatalf("expected exit 0, got exit=%v code=%d", exit, code)
	}
}

func TestDispatchCLIUnknownCommand(t *testing.T) {
	exit, code := dispatchCLI([]string{"docklog", "not-a-command"})
	if !exit || code != 1 {
		t.Fatalf("expected exit 1, got exit=%v code=%d", exit, code)
	}
}

func TestApplyRunModes(t *testing.T) {
	t.Setenv("DOCKLOG_MODE", "")
	t.Setenv("DOCKLOG_AGENT_ONLY", "")

	applyRunMode("agent-only")
	if serveFrontend {
		t.Fatal("agent-only should not serve frontend")
	}
	if os.Getenv("DOCKLOG_AGENT_ONLY") != "true" {
		t.Fatal("expected DOCKLOG_AGENT_ONLY=true")
	}

	applyRunMode("agent")
	if !serveFrontend {
		t.Fatal("agent should serve frontend")
	}

	applyRunMode("server")
	if !serveFrontend {
		t.Fatal("server should serve frontend")
	}
}

func TestPrintConfig(t *testing.T) {
	out := captureOutput(printConfig)
	if !bytes.Contains([]byte(out), []byte("DockLog configuration")) {
		t.Fatalf("unexpected config output: %q", out)
	}
}
