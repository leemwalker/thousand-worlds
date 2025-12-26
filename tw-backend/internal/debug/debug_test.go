package debug

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestFlags(t *testing.T) {
	SetFlags(None)
	if Is(Perf) {
		t.Error("Perf should be disabled by default")
	}

	Enable(Perf)
	if !Is(Perf) {
		t.Error("Perf should be enabled after Enable()")
	}
	if Is(Logic) {
		t.Error("Logic should still be disabled")
	}

	SetFlags(All)
	if !Is(Perf) || !Is(Logic) || !Is(Geology) {
		t.Error("All flags should be enabled")
	}

	Disable(Perf)
	if Is(Perf) {
		t.Error("Perf should be disabled after Disable()")
	}
	if !Is(Logic) {
		t.Error("Logic should remain enabled")
	}
}

func TestLog(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	SetFlags(Perf)

	Log(Perf, "Perf Check")
	Log(Logic, "Logic Check")

	output := buf.String()
	if !strings.Contains(output, "Perf Check") {
		t.Error("Should have logged Perf message")
	}
	if strings.Contains(output, "Logic Check") {
		t.Error("Should NOT have logged Logic message")
	}
}
