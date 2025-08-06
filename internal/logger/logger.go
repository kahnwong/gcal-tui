package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Configure zerolog with console writer for better readability
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

// GetLogger returns the configured zerolog logger
func GetLogger() zerolog.Logger {
	return log.Logger
}
