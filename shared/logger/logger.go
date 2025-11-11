package logger

import (
	"io"
	"log"
	"os"
)

// Logger is a structured logger
type Logger struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// New creates a new logger
func New(serviceName string) *Logger {
	prefix := "[" + serviceName + "] "

	return &Logger{
		infoLog:  log.New(os.Stdout, prefix+"INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(os.Stderr, prefix+"ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLog: log.New(os.Stdout, prefix+"DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// NewWithWriter creates a logger with custom writers
func NewWithWriter(serviceName string, infoWriter, errorWriter io.Writer) *Logger {
	prefix := "[" + serviceName + "] "

	return &Logger{
		infoLog:  log.New(infoWriter, prefix+"INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLog: log.New(errorWriter, prefix+"ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLog: log.New(infoWriter, prefix+"DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *Logger) Info(v ...interface{}) {
	l.infoLog.Println(v...)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.infoLog.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(v ...interface{}) {
	l.errorLog.Println(v...)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
}

// Debug logs a debug message
func (l *Logger) Debug(v ...interface{}) {
	l.debugLog.Println(v...)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.debugLog.Printf(format, v...)
}

// Fatal logs a fatal error and exits
func (l *Logger) Fatal(v ...interface{}) {
	l.errorLog.Fatal(v...)
}

// Fatalf logs a formatted fatal error and exits
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.errorLog.Fatalf(format, v...)
}
