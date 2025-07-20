package common

import (
	"github.com/neilberkman/clippy/internal/log"
)

// SetupLogger creates a new logger with the given verbose and debug settings
func SetupLogger(verbose, debug bool) *log.Logger {
	return log.New(log.Config{
		Verbose: verbose || debug,
		Debug:   debug,
	})
}
