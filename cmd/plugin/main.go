package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	console = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	logger = zerolog.New(os.Stderr).Output(console).With().Timestamp().Caller().Logger()
)

func main() {
	ctx := context.Background()
	if err := executeContext(ctx); err != nil {
		logger.Fatal().Err(err).Msg("failed to execute command")
	}
}
