package main

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	console := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}
	log.Logger = log.Logger.Output(console).With().Caller().Logger()
}

func main() {
	ctx := context.Background()
	if err := executeContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
