package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ihcsim/k8s-dra/cmd/flags"
	draclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	gpukubeletplugin "github.com/ihcsim/k8s-dra/pkg/drivers/gpu/kubelet"
	"golang.org/x/exp/rand"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/dynamic-resource-allocation/kubeletplugin"
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

	if err := viper.BindEnv("node-name", "NODE_NAME"); err != nil {
		log.Fatal().Err(err).Msg("failed to bind env vars")
	}
}

func executeContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func run(ctx context.Context) error {
	var (
		kubeconfig       = viper.GetString("kubeconfig")
		cdiRoot          = viper.GetString("cdi-root")
		namespace        = viper.GetString("namespace")
		nodeName         = viper.GetString("node-name")
		maxAvailableGPU  = viper.GetInt("max-available-gpu")
		randAvailableGPU = rand.Intn(maxAvailableGPU)
	)

	if err := os.MkdirAll(pluginPath, 0750); err != nil {
		return err
	}

	if err := os.MkdirAll(cdiRoot, 0750); err != nil {
		return err
	}

	draClientSets, err := draClientSets(kubeconfig)
	if err != nil {
		return err
	}

	log.Info().Msgf("Starting DRA node server with %d available GPUs", randAvailableGPU)
	nodeServer, err := gpukubeletplugin.NewNodeServer(ctx, draClientSets, cdiRoot, namespace, nodeName, randAvailableGPU, log.Logger)
	if err != nil {
		return err
	}

	p, err := kubeletplugin.Start(
		nodeServer,
		kubeletplugin.DriverName(driverName),
		kubeletplugin.RegistrarSocketPath(pluginRegistrationPath),
		kubeletplugin.PluginSocketPath(pluginSocketPath),
		kubeletplugin.KubeletPluginSocketPath(pluginSocketPath),
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

func kubeConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return rest.InClusterConfig()
}

func draClientSets(kubeconfigPath string) (draclientset.Interface, error) {
	kubecfg, err := kubeConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return draclientset.NewForConfig(kubecfg)
}
