package log

import (
	"fmt"
	"os"
)

// Config holds logging configuration
type Config struct {
	Verbose bool
	Debug   bool
}

// Logger provides logging functionality
type Logger struct {
	config Config
}

// New creates a new logger with the given configuration
func New(config Config) *Logger {
	return &Logger{config: config}
}

// Error prints an error message and exits
func (l *Logger) Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

// Verbose prints a message if verbose mode is enabled
func (l *Logger) Verbose(format string, args ...interface{}) {
	if l.config.Verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// Debug prints a message if debug mode is enabled
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.config.Debug {
		fmt.Printf("DEBUG: "+format+"\n", args...)
	}
}

// Warning prints a warning message to stderr if verbose mode is enabled
func (l *Logger) Warning(format string, args ...interface{}) {
	if l.config.Verbose {
		fmt.Fprintf(os.Stderr, "Warning: "+format+"\n", args...)
	}
}

// Print always prints a message (used for required output)
func (l *Logger) Print(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// PrintErr always prints to stderr (used for required errors/warnings)
func (l *Logger) PrintErr(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
