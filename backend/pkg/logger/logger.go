package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a configured zerolog.Logger.
// In development: pretty console output.
// In production: structured JSON output.
func New(env string) zerolog.Logger {
	var w io.Writer

	if env == "production" {
		w = os.Stdout
	} else {
		w = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	return zerolog.New(w).
		With().
		Timestamp().
		Caller().
		Logger()
}
