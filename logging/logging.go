package logging

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/UCCNetworkingSociety/Netsoc-Discord-Bot/config"
)

// Logger wraps up all your logging needs in one struct
type Logger struct {
	logFiles      []*os.File
	infol, errorl *log.Logger
}

// conextKey is the type for the key for which loggers will be associated with
// within a context. It is unexported to prevent collisions with other
// context keys.
type conextKey int

// loggerContextKey is the key-value to which loggers will be associated with
// within a context.
const loggerContextKey conextKey = 0

// New creates a new logger which logs to the given log file.
// Note: you must defer a call to the logger's "Close" method.
func New() (*Logger, error) {
	conf := config.GetConfig()

	infolf, err := os.OpenFile(conf.LogFiles.InfoLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open info log file 'info.log': %s", err)
	}
	errorlf, err := os.OpenFile(conf.LogFiles.ErrorLog, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log file 'error.log': %s", err)
	}
	return &Logger{
		logFiles: []*os.File{infolf, errorlf},
		infol:    log.New(infolf, "INFO:  ", log.Ldate|log.Ltime|log.Lshortfile),
		errorl:   log.New(errorlf, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

// Infof outputs log messages to the info log only.
func (l *Logger) Infof(fmtMsg string, opts ...interface{}) {
	if !strings.HasSuffix(fmtMsg, "\n") {
		fmtMsg += "\n"
	}
	fmt.Printf("[%s] Info: %s", time.Now(), fmt.Sprintf(fmtMsg, opts...))
	l.infol.Printf(fmtMsg, opts...)
}

// Errorf outpus log messages to the error log and lower severities.
func (l *Logger) Errorf(fmtMsg string, opts ...interface{}) {
	if !strings.HasSuffix(fmtMsg, "\n") {
		fmtMsg += "\n"
	}
	fmt.Printf("[%s] Error: %s", time.Now(), fmt.Sprintf(fmtMsg, opts...))
	l.infol.Printf(fmtMsg, opts...)
	l.errorl.Printf(fmtMsg, opts...)
}

// Close closes the logger's log file.
func (l *Logger) Close() {
	for _, lf := range l.logFiles {
		lf.Close()
	}
}

// FromContext will return a Logger which is associated with the given context.
// If there is no logger within the context, then nil is returned and ok is false.
func FromContext(ctx context.Context) (*Logger, bool) {
	l, ok := ctx.Value(loggerContextKey).(*Logger)
	return l, ok
}

// NewContext associates the given logger with the given context and returns the
// new child context.
func NewContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}
