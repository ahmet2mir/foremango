package foreman

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

// recordingLogger is a minimal Logger that records every call it receives,
// standing in for an external logging package (zap, logrus, slog, ...).
type recordingLogger struct {
	calls []string
}

func (r *recordingLogger) Tracef(format string, a ...interface{}) { r.calls = append(r.calls, "TRACE") }
func (r *recordingLogger) Debugf(format string, a ...interface{}) { r.calls = append(r.calls, "DEBUG") }
func (r *recordingLogger) Infof(format string, a ...interface{})  { r.calls = append(r.calls, "INFO") }
func (r *recordingLogger) Warningf(format string, a ...interface{}) {
	r.calls = append(r.calls, "WARNING")
}
func (r *recordingLogger) Errorf(format string, a ...interface{}) { r.calls = append(r.calls, "ERROR") }

// Ensures a Client built with ClientConfig.Logger set routes its internal
// logging (c.log.Xxx) through that Logger instead of the built-in default.
func TestClientConfig_LoggerOverride(t *testing.T) {
	rec := &recordingLogger{}
	client := NewClient(Server{}, ClientCredentials{}, ClientConfig{Logger: rec})

	// NewRequestWithContext logs via c.log.Debugf on the happy path and
	// c.log.Errorf when given an invalid HTTP method.
	if _, err := client.NewRequestWithContext(context.TODO(), "BOGUS", "/foo", nil); err == nil {
		t.Fatalf("expected an error for an invalid HTTP method")
	}

	if len(rec.calls) == 0 {
		t.Fatalf("expected the client's configured Logger to receive at least one call, got none")
	}
	for _, call := range rec.calls {
		if call != "DEBUG" && call != "ERROR" {
			t.Fatalf("unexpected call recorded on the configured Logger: %s", call)
		}
	}
}

// Ensures a Client without ClientConfig.Logger set falls back to the
// built-in default logger, and that one client's custom Logger is never
// used by another client that didn't configure one.
func TestClientConfig_LoggerDefaultWhenUnset(t *testing.T) {
	rec := &recordingLogger{}
	_ = NewClient(Server{}, ClientCredentials{}, ClientConfig{Logger: rec})
	withoutCustom := NewClient(Server{}, ClientCredentials{}, ClientConfig{})

	_, _ = withoutCustom.NewRequestWithContext(context.TODO(), "BOGUS", "/foo", nil)
	if len(rec.calls) != 0 {
		t.Fatalf("a client without ClientConfig.Logger must not log through another client's configured Logger")
	}
}

func TestLogLevel_String(t *testing.T) {
	cases := map[LogLevel]string{
		LevelTrace:   "TRACE",
		LevelDebug:   "DEBUG",
		LevelInfo:    "INFO",
		LevelWarning: "WARNING",
		LevelError:   "ERROR",
		LevelNone:    "NONE",
		LogLevel(99): "",
		LogLevel(-1): "",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("LogLevel(%d).String() = %q, want %q", int(level), got, want)
		}
	}
}

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in     string
		want   LogLevel
		wantOk bool
	}{
		{"trace", LevelTrace, true},
		{" TRACE ", LevelTrace, true},
		{"debug", LevelDebug, true},
		{"info", LevelInfo, true},
		{"warning", LevelWarning, true},
		{"warn", LevelWarning, true},
		{"error", LevelError, true},
		{"none", LevelNone, true},
		{"off", LevelNone, true},
		{"nonsense", LevelInfo, false},
	}
	for _, c := range cases {
		got, ok := parseLevel(c.in)
		if got != c.want || ok != c.wantOk {
			t.Errorf("parseLevel(%q) = (%v, %v), want (%v, %v)", c.in, got, ok, c.want, c.wantOk)
		}
	}
}

// Exercises the package-level logging functions (used by code with no
// *Client at hand) against the shared default logger, restoring its level
// and output afterwards so this test doesn't leak state into others.
func TestPackageLevelLogging(t *testing.T) {
	prevLevel := Level()
	defer SetLevel(prevLevel)
	defer SetOutput(io.Discard) // matches TestMain's default for every other test

	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(LevelTrace)

	Tracef("trace %d", 1)
	Debug("debug %d", 2)
	Infof("info %d", 3)
	Warningf("warning %d", 4)
	Errorf("error %d", 5)
	Fatalf("fatal %d", 6)
	Fatal("fatal-single")
	TraceFunctionCall()

	out := buf.String()
	for _, want := range []string{
		"[TRACE] trace 1",
		"[DEBUG] debug 2",
		"[INFO] info 3",
		"[WARNING] warning 4",
		"[ERROR] error 5",
		"[ERROR] FATAL: fatal 6",
		"[ERROR] FATAL: fatal-single",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected log output to contain %q, got:\n%s", want, out)
		}
	}
}
