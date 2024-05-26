package main

import (
	"context"

	"github.com/ihcsim/k8s-dra/cmd/flags"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dra-plugin",
		Short: "dra-plugin implements a DRA kubelet plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
)

func init() {
	rootCmd.PersistentFlags().AddFlagSet(flags.NewK8sFlags())
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("failed to bind flags")
	}
}

func executeContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func run(ctx context.Context) error {
	return nil
}
