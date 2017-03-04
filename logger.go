package copier

import "context"

const (
	// ERROR logger severity.
	ERROR = iota
	// WARN logger severity.
	WARN
	// INFO logger severity.
	INFO
	// DEBUG logger severity.
	DEBUG
)

type (
	// LogEntry contains detials about an Exec log.
	LogEntry struct {
		Severity    int
		Status      string
		FilepathSrc string
		FilepathDst string
	}

	// A Logger is an LogEntry centralizer.
	Logger struct {
		context.Context
		C chan LogEntry
	}
)

// NewLogger returns a new Logger.
func NewLogger() *Logger {
	return NewLoggerWithContext(context.Background())
}

// NewLoggerWithContext returns a new Logger with the given ctx.
func NewLoggerWithContext(ctx context.Context) *Logger {
	return &Logger{
		Context: ctx,
		C:       make(chan LogEntry, 8),
	}
}
