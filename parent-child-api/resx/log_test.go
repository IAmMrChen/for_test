package resx

import (
	"bytes"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestLogWriterUsesStdoutOutsideLocal(t *testing.T) {
	writer := newLogWriter()
	if writer != os.Stdout {
		t.Fatalf("prod log writer = %T, want os.Stdout", writer)
	}
}

func TestSetupLogUsesDefaultLoggerForRootAndJob(t *testing.T) {
	previousDefault := slog.Default()
	previousLog := Log
	defer func() {
		slog.SetDefault(previousDefault)
		Log = previousLog
	}()

	var buf bytes.Buffer
	setupLog("prod", &buf)

	if Log.Root != slog.Default() {
		t.Fatal("root logger should reuse slog.Default")
	}

	Log.Root.Info("root message")
	Log.Job.Info("job message")

	output := buf.String()
	if !strings.Contains(output, "msg=root message") {
		t.Fatalf("root logger did not write through configured writer: %q", output)
	}
	if !strings.Contains(output, "msg=job message") {
		t.Fatalf("job logger did not write through configured writer: %q", output)
	}
	if !strings.Contains(strings.ToLower(output), "logger=job") {
		t.Fatalf("job logger should include logger=job field: %q", output)
	}
}

func TestSetupLogDevelopmentEnablesDebugAndSource(t *testing.T) {
	previousDefault := slog.Default()
	previousLog := Log
	defer func() {
		slog.SetDefault(previousDefault)
		Log = previousLog
	}()

	var buf bytes.Buffer
	setupLog("dev", &buf)

	Log.Root.Debug("debug message")

	output := buf.String()
	if !strings.Contains(output, "msg=debug message") {
		t.Fatalf("dev logger should emit debug logs: %q", output)
	}
	if !strings.Contains(output, "source=") {
		t.Fatalf("dev logger should include source location: %q", output)
	}
}
