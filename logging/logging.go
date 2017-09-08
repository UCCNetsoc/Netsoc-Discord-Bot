package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// Logger wraps up all your logging needs in one struct
type Logger struct {
	logFiles      []*os.File
	infol, errorl *log.Logger
}

// New creates a new logger which logs to the given log file.
// Note: you must defer a call to the logger's "Close" method.
func New() (*Logger, error) {
	infolf, err := os.OpenFile("info.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open info log file 'info.log': %s", err)
	}
	errorlf, err := os.OpenFile("error.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log file 'error.log': %s", err)
	}
	return &Logger{
		logFiles: []*os.File{infolf, errorlf},
		infol:    log.New(infolf, "INFO:  ", log.Ldate|log.Ltime),
		errorl:   log.New(errorlf, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

// Infof outputs log messages to the info log only.
func (l *Logger) Infof(fmtMsg string, opts ...interface{}) {
	if !strings.HasSuffix(fmtMsg, "\n") {
		fmtMsg += "\n"
	}
	fmt.Printf("I: %s", fmt.Sprintf(fmtMsg, opts...))
	l.infol.Printf(fmtMsg, opts...)
}

// Errorf outpus log messages to the error log and lower severities.
func (l *Logger) Errorf(fmtMsg string, opts ...interface{}) {
	if !strings.HasSuffix(fmtMsg, "\n") {
		fmtMsg += "\n"
	}
	fmt.Printf("E: %s", fmt.Sprintf(fmtMsg, opts...))
	l.infol.Printf(fmtMsg, opts...)
	l.errorl.Printf(fmtMsg, opts...)
}

// Close closes the logger's log file.
func (l *Logger) Close() {
	for _, lf := range l.logFiles {
		lf.Close()
	}
}
