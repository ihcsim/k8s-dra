package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ihcsim/k8s-dra/cmd/flags"
	"github.com/ihcsim/k8s-dra/pkg/drivers/gpu"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	plugin "k8s.io/dynamic-resource-allocation/kubeletplugin"
)

const driverName = "driver.resources.ihcsim"

var (
	rootCmd = &cobra.Command{
		Use:   "dra-plugin",
		Short: "dra-plugin implements a DRA kubelet plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}

	pluginRegistrationPath = "/var/lib/kubelet/plugins_registry/" + driverName + ".sock"
	pluginPath             = "/var/lib/kubelet/plugins/" + driverName
	pluginSocketPath       = pluginPath + "/plugin.sock"
)

func init() {
	rootCmd.PersistentFlags().AddFlagSet(flags.NewK8sFlags())
	rootCmd.PersistentFlags().AddFlagSet(flags.NewPluginFlags())

	if err := rootCmd.MarkPersistentFlagRequired("cdi-root"); err != nil {
		log.Fatal().Err(err).Msg("failed to mark flags as required")
	}

	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("failed to bind flags")
	}
}

func executeContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func run(ctx context.Context) error {
	if err := os.MkdirAll(pluginPath, 0750); err != nil {
		return err
	}

	cdiRoot := viper.GetString("cdi-root")
	if err := os.MkdirAll(cdiRoot, 0750); err != nil {
		return err
	}

	nodeServer := gpu.NewNodeServer(ctx)
	p, err := plugin.Start(
		nodeServer,
		plugin.DriverName(driverName),
		plugin.RegistrarSocketPath(pluginRegistrationPath),
		plugin.PluginSocketPath(pluginSocketPath),
		plugin.KubeletPluginSocketPath(pluginSocketPath),
	)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan

	p.Stop()
	return nodeServer.Shutdown(ctx)
}
