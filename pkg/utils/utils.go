// Package utils provides small internal helpers for the api package,
// chiefly a leveled logger used throughout pkg/api for diagnostics.
//
// It has no dependency outside the Go standard library by design: pkg/api
// is meant to be usable as a plain Go library, not just from inside a
// Terraform provider process.
package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

// LogLevel is the verbosity threshold of the package logger. Levels are
// ordered from most to least verbose; a logger set to a given level emits
// messages at that level and above.
type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarning
	LevelError
	LevelNone
)

// String returns the human-readable name of the level, or "" for an
// out-of-range value.
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelNone:
		return "NONE"
	default:
		return ""
	}
}

var (
	mu    sync.Mutex
	level = LevelInfo
	out   = log.New(os.Stderr, "", log.LstdFlags)
)

func init() {
	if v := os.Getenv("FOREMANGO_LOG_LEVEL"); v != "" {
		if l, ok := parseLevel(v); ok {
			level = l
		}
	}
}

func parseLevel(s string) (LogLevel, bool) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "TRACE":
		return LevelTrace, true
	case "DEBUG":
		return LevelDebug, true
	case "INFO":
		return LevelInfo, true
	case "WARNING", "WARN":
		return LevelWarning, true
	case "ERROR":
		return LevelError, true
	case "NONE", "OFF":
		return LevelNone, true
	default:
		return LevelInfo, false
	}
}

// SetLevel sets the minimum level of messages written to the log output.
// The default level is LevelInfo, or whatever FOREMANGO_LOG_LEVEL was set
// to at process start.
func SetLevel(l LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	level = l
}

// Level returns the current minimum log level.
func Level() LogLevel {
	mu.Lock()
	defer mu.Unlock()
	return level
}

// SetOutput redirects log output to w. Defaults to os.Stderr.
func SetOutput(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	out.SetOutput(w)
}

func logf(l LogLevel, format string, a ...interface{}) {
	mu.Lock()
	threshold := level
	mu.Unlock()
	if l < threshold {
		return
	}
	out.Printf("[%s] %s", l, fmt.Sprintf(format, a...))
}

// Tracef logs a message at TRACE level - the flow of execution, such as
// function enter/exit notifications. See TraceFunctionCall for a shorthand.
func Tracef(format string, a ...interface{}) {
	logf(LevelTrace, format, a...)
}

// Debugf logs a message at DEBUG level - intermediate values and
// calculations useful when stepping through a problem.
func Debugf(format string, a ...interface{}) {
	logf(LevelDebug, format, a...)
}

// Debug is an alias for Debugf, kept because existing call sites in this
// module use both spellings interchangeably.
func Debug(format string, a ...interface{}) {
	logf(LevelDebug, format, a...)
}

// Infof logs a message at INFO level.
func Infof(format string, a ...interface{}) {
	logf(LevelInfo, format, a...)
}

// Warningf logs a message at WARNING level.
func Warningf(format string, a ...interface{}) {
	logf(LevelWarning, format, a...)
}

// Errorf logs a message at ERROR level.
func Errorf(format string, a ...interface{}) {
	logf(LevelError, format, a...)
}

// Fatalf logs a message at ERROR level, prefixed with "FATAL".
//
// NOTE: unlike the stdlib log.Fatalf, this does NOT call os.Exit. pkg/api is
// a library, and a library must never terminate its host process out from
// under the caller just because one API response was malformed - the
// caller's error handling should decide what to do. Call sites in pkg/api
// that invoke this on bad input generally have a bug (they should return an
// error instead); logging loudly at ERROR level is the safe stand-in.
func Fatalf(format string, a ...interface{}) {
	logf(LevelError, "FATAL: "+format, a...)
}

// Fatal is an alias for Fatalf for a single value. See Fatalf.
func Fatal(a interface{}) {
	Fatalf("%v", a)
}

// TraceFunctionCall logs, at TRACE level, the name, file and line of the
// function that called it.
func TraceFunctionCall() {
	// Skip the (relatively expensive) runtime.Caller lookup entirely when
	// TRACE isn't even going to be emitted.
	if Level() > LevelTrace {
		return
	}
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		Tracef("function call")
		return
	}
	name := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		name = fn.Name()
	}
	Tracef("%s (%s:%d)", name, file, line)
}
