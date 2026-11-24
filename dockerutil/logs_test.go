package dockerutil

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestStreamLogs_Tty(t *testing.T) {
	raw := "2026-11-24T12:00:00Z line 1\n2026-11-24T12:00:01Z line 2\n"
	buf := bytes.NewBufferString(raw)

	var lines []string
	err := StreamLogs(buf, true, func(payload []byte) error {
		lines = append(lines, string(payload))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "2026-11-24T12:00:00Z line 1\n" {
		t.Errorf("unexpected line 0: %q", lines[0])
	}
}

func TestStreamLogs_Multiplexed(t *testing.T) {
	var buf bytes.Buffer

	// Frame 1: stdout, 5 bytes "hello"
	h1 := []byte{1, 0, 0, 0, 0, 0, 0, 5}
	buf.Write(h1)
	buf.WriteString("hello")

	// Frame 2: stderr, 5 bytes "world"
	h2 := []byte{2, 0, 0, 0, 0, 0, 0, 5}
	buf.Write(h2)
	buf.WriteString("world")

	var payloads []string
	err := StreamLogs(&buf, false, func(payload []byte) error {
		payloads = append(payloads, string(payload))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 2 {
		t.Fatalf("expected 2 payloads, got %d", len(payloads))
	}
	if payloads[0] != "hello" || payloads[1] != "world" {
		t.Errorf("unexpected payloads: %v", payloads)
	}
}

func TestStreamLogs_MalformedRawTextFallback(t *testing.T) {
	// Raw text that starts with ASCII characters like "2026-11-24T12:00:00Z app started\n"
	// Previously, byte 4-7 ('-11-') was treated as uint32 size 758,206,765 bytes!
	raw := "2026-11-24T12:00:00Z app started\nsecond line\n"
	buf := bytes.NewBufferString(raw)

	var payloads []string
	err := StreamLogs(buf, false, func(payload []byte) error {
		payloads = append(payloads, string(payload))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(payloads) != 2 {
		t.Fatalf("expected 2 payloads via fallback, got %d", len(payloads))
	}
	if payloads[0] != "2026-11-24T12:00:00Z app started\n" {
		t.Errorf("unexpected payload 0: %q", payloads[0])
	}
	if payloads[1] != "second line\n" {
		t.Errorf("unexpected payload 1: %q", payloads[1])
	}
}

func TestStreamLogs_OversizedFrameSafety(t *testing.T) {
	var buf bytes.Buffer
	// Frame claiming 500MB size
	hdr := make([]byte, 8)
	hdr[0] = 1 // stdout
	binary.BigEndian.PutUint32(hdr[4:], 500*1024*1024)
	buf.Write(hdr)

	called := false
	err := StreamLogs(&buf, false, func(payload []byte) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil on safety cap abort, got %v", err)
	}
	if called {
		t.Errorf("handler should not have been called for 500MB frame")
	}
}

