package logger

import (
	"log"
	"os"
)

// Logger handles logging for the application
type Logger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
}

// NewLogger creates a new Logger
func New() *Logger {
	return &Logger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

// Info logs informational messages to stdout
func (l *Logger) Info(format string, v ...any) {
	l.infoLogger.Printf(format, v...)
}

// Error logs error messages to stderr
func (l *Logger) Error(format string, v ...any) {
	l.errorLogger.Printf(format, v...)
}
