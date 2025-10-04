package utils

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents logging levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var currentLevel LogLevel

func init() {
	// Initialize log level from environment
	level := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	switch level {
	case "debug":
		currentLevel = DEBUG
	case "info", "": // Default to INFO
		currentLevel = INFO
	case "warn", "warning":
		currentLevel = WARN
	case "error":
		currentLevel = ERROR
	default:
		currentLevel = INFO
	}
}

// Debug logs a debug message (only if LOG_LEVEL=debug)
func Debug(format string, v ...interface{}) {
	if currentLevel <= DEBUG {
		log.Printf("DEBUG: "+format, v...)
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	if currentLevel <= INFO {
		log.Printf("INFO: "+format, v...)
	}
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	if currentLevel <= WARN {
		log.Printf("WARN: "+format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if currentLevel <= ERROR {
		log.Printf("ERROR: "+format, v...)
	}
}
