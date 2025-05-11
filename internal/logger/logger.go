package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func New(verbose bool) zerolog.Logger {
	level := zerolog.InfoLevel
	if verbose {
		level = zerolog.DebugLevel
	}
	return zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Logger().
		Level(level)
}
