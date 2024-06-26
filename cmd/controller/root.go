package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/ihcsim/k8s-dra/cmd/flags"
	draclientset "github.com/ihcsim/k8s-dra/pkg/apis/clientset/versioned"
	"github.com/ihcsim/k8s-dra/pkg/drivers/gpu"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/client-go/informers"
	coreclientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/dynamic-resource-allocation/controller"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dra-ctrl",
		Short: "dra-ctrl implements a Kubernetes DRA driver controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context())
		},
	}
)

func init() {
	rootCmd.PersistentFlags().AddFlagSet(flags.NewK8sFlags())
	rootCmd.PersistentFlags().AddFlagSet(flags.NewControllerFlags())
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		log.Fatal().Err(err).Msg("failed to bind flags")
	}
}

func executeContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func run(ctx context.Context) error {
	var (
		kubeconfig  = viper.GetString("kubeconfig")
		workerCount = viper.GetInt("workers")
		qps         = viper.GetFloat64("api-qps")
		burst       = viper.GetFloat64("api-burst")

		metricsPort = viper.GetInt("metrics-port")
		metricsPath = viper.GetString("metrics-path")
		pprofPort   = viper.GetInt("pprof-port")
		pprofPath   = "/debug/pprof/"

		namespace = viper.GetString("namespace")
	)

	go func() {
		s := http.NewServeMux()
		s.HandleFunc(pprofPath, pprof.Index)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", pprofPort), s); err != nil {
			log.Warn().Err(err).Msg("failed to start pprof server")
		}
	}()

	go func() {
		s := http.NewServeMux()
		s.Handle(fmt.Sprintf("/%s", metricsPath), promhttp.Handler())
		if err := http.ListenAndServe(fmt.Sprintf(":%d", metricsPort), s); err != nil {
			log.Warn().Err(err).Msg("failed to start metrics server")
		}
	}()

	log.Info().
		Int("workers", workerCount).
		Float64("qps", qps).
		Float64("burst", burst).
		Str("metrics", fmt.Sprintf("/%s:%d", metricsPath, metricsPort)).
		Str("pprof", fmt.Sprintf("%s:%d", pprofPath, pprofPort)).
		Send()

	coreClientSets, err := coreClientSets(kubeconfig)
	if err != nil {
		return err
	}

	draClientSets, err := draClientSets(kubeconfig)
	if err != nil {
		return err
	}

	var (
		resync          = time.Minute * 10
		informerFactory = informers.NewSharedInformerFactory(coreClientSets, resync)
	)
	informerFactory.Start(ctx.Done())

	driverLog := log.Logger.With().Str("namespace", namespace).Logger()
	driver, err := gpu.NewDriver(draClientSets, namespace, driverLog)
	if err != nil {
		return err
	}

	log.Info().Str("driver", driver.GetName()).Msg("starting driver controller")
	ctrl := controller.New(ctx, driver.GetName(), driver, coreClientSets, informerFactory)
	ctrl.Run(workerCount)
	return nil
}

func kubeConfig(kubeconfigPath string) (*rest.Config, error) {
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return rest.InClusterConfig()
}

func coreClientSets(kubeconfigPath string) (coreclientset.Interface, error) {
	kubecfg, err := kubeConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return coreclientset.NewForConfig(kubecfg)
}

func draClientSets(kubeconfigPath string) (draclientset.Interface, error) {
	kubecfg, err := kubeConfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	return draclientset.NewForConfig(kubecfg)
}
