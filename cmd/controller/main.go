package main

import (
	"context"

	"github.com/ihcsim/k8s-dra/cmd"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := context.Background()
	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to execute command")
	}
}
