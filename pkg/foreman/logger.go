// The logger used by this package. By default it's a small stdlib-based
// leveled logger (defaultLogger below); every Client gets its own Logger at
// construction time (see Client.log in client.go) - override it per-client
// via ClientConfig.Logger, e.g. to route a given client's diagnostics
// through zap, logrus, slog, or anything else that implements the
// five-method Logger interface.
//
// A handful of call sites that run before any Client exists, or that have
// no Client to hand (a resource type's UnmarshalJSON, for instance), use
// the package-level functions below instead; those always go through
// defaultLogger, which SetLevel/SetOutput configure directly.
package foreman

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
)

// LogLevel is the verbosity threshold of a Logger. Levels are ordered from
// most to least verbose; a logger set to a given level emits messages at
// that level and above.
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

// Logger is the logging interface Client calls into (as c.log) and that the
// package-level functions below delegate to by default. Implement it in
// your own package and set it on ClientConfig.Logger to take over a given
// client's logging - e.g. to route it through zap, logrus, slog, or
// straight to /dev/null.
type Logger interface {
	Tracef(format string, a ...interface{})
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

// defaultLogger is the built-in stdlib-based Logger. NewClient uses it for
// any Client whose ClientConfig.Logger is nil, and the package-level
// functions below (for call sites with no Client at hand) always use it.
var defaultLogger = newStdLogger()

func init() {
	if v := os.Getenv("FOREMANGO_LOG_LEVEL"); v != "" {
		if l, ok := parseLevel(v); ok {
			defaultLogger.SetLevel(l)
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

// SetLevel sets the minimum level of messages written by defaultLogger - the
// logger every Client uses unless ClientConfig.Logger overrides it, and the
// one the package-level Tracef/Debugf/... functions always use.
//
// The default level is LevelInfo, or whatever FOREMANGO_LOG_LEVEL was set
// to at process start.
func SetLevel(l LogLevel) {
	defaultLogger.SetLevel(l)
}

// Level returns defaultLogger's current minimum log level.
func Level() LogLevel {
	return defaultLogger.Level()
}

// SetOutput redirects defaultLogger's output to w. Defaults to os.Stderr.
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Tracef logs a message at TRACE level through defaultLogger - the flow of
// execution, such as function enter/exit notifications. See
// TraceFunctionCall for a shorthand. Client-bound code should prefer
// c.log.Tracef instead; this is for call sites with no Client at hand.
func Tracef(format string, a ...interface{}) {
	defaultLogger.Tracef(format, a...)
}

// Debugf logs a message at DEBUG level through defaultLogger - intermediate
// values and calculations useful when stepping through a problem. See the
// Tracef note above.
func Debugf(format string, a ...interface{}) {
	defaultLogger.Debugf(format, a...)
}

// Debug is an alias for Debugf, kept because existing call sites in this
// module use both spellings interchangeably.
func Debug(format string, a ...interface{}) {
	defaultLogger.Debugf(format, a...)
}

// Infof logs a message at INFO level through defaultLogger. See the Tracef
// note above.
func Infof(format string, a ...interface{}) {
	defaultLogger.Infof(format, a...)
}

// Warningf logs a message at WARNING level through defaultLogger. See the
// Tracef note above.
func Warningf(format string, a ...interface{}) {
	defaultLogger.Warningf(format, a...)
}

// Errorf logs a message at ERROR level through defaultLogger. See the
// Tracef note above.
func Errorf(format string, a ...interface{}) {
	defaultLogger.Errorf(format, a...)
}

// Fatalf logs a message at ERROR level through defaultLogger, prefixed with
// "FATAL".
//
// NOTE: unlike the stdlib log.Fatalf, this does NOT call os.Exit. foremango
// is a library, and a library must never terminate its host process out
// from under the caller just because one API response was malformed - the
// caller's error handling should decide what to do. Call sites that invoke
// this on bad input generally have a bug (they should return an error
// instead); logging loudly at ERROR level is the safe stand-in.
func Fatalf(format string, a ...interface{}) {
	defaultLogger.Errorf("FATAL: "+format, a...)
}

// Fatal is an alias for Fatalf for a single value. See Fatalf.
func Fatal(a interface{}) {
	Fatalf("%v", a)
}

// TraceFunctionCall logs, at TRACE level through defaultLogger, the name,
// file and line of the function that called it. Used uniformly everywhere
// in this module, Client-bound code included - it's a generic debugging aid
// independent of any one Client's configured logger.
func TraceFunctionCall() {
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

// ----------------------------------------------------------------------------
// Built-in default Logger
// ----------------------------------------------------------------------------

// stdLogger is the default Logger: a small level-filtered wrapper around the
// standard library's log.Logger.
type stdLogger struct {
	mu    sync.Mutex
	level LogLevel
	out   *log.Logger
}

func newStdLogger() *stdLogger {
	return &stdLogger{
		level: LevelInfo,
		out:   log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (s *stdLogger) SetLevel(l LogLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.level = l
}

func (s *stdLogger) Level() LogLevel {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.level
}

func (s *stdLogger) SetOutput(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.out.SetOutput(w)
}

func (s *stdLogger) logf(l LogLevel, format string, a ...interface{}) {
	s.mu.Lock()
	threshold := s.level
	s.mu.Unlock()
	if l < threshold {
		return
	}
	s.out.Printf("[%s] %s", l, fmt.Sprintf(format, a...))
}

func (s *stdLogger) Tracef(format string, a ...interface{})   { s.logf(LevelTrace, format, a...) }
func (s *stdLogger) Debugf(format string, a ...interface{})   { s.logf(LevelDebug, format, a...) }
func (s *stdLogger) Infof(format string, a ...interface{})    { s.logf(LevelInfo, format, a...) }
func (s *stdLogger) Warningf(format string, a ...interface{}) { s.logf(LevelWarning, format, a...) }
func (s *stdLogger) Errorf(format string, a ...interface{})   { s.logf(LevelError, format, a...) }
