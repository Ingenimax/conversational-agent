package logger

import (
	"os"
	"sync"

	"github.com/rs/zerolog"
)

var (
	once   sync.Once
	logger zerolog.Logger
)

// GetLogger initializes and returns a shared logger instance
func GetLogger() zerolog.Logger {
	once.Do(func() {
		output := zerolog.ConsoleWriter{Out: os.Stdout}
		logger = zerolog.New(output).With().Timestamp().Logger()
	})
	return logger
}
