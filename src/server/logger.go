package main

import (
	"github.com/rs/zerolog"
	"os"
)

func Logger() zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	return logger
}
