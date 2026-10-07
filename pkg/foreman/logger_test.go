package foreman

import (
	"context"
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
