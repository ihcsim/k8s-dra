package flags

import (
	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/clientcmd"
)

func NewK8sFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("k8s", pflag.ExitOnError)
	flags.String("kubeconfig", clientcmd.RecommendedHomeFile, "Path to the kubeconfig file")
	flags.Float64("api-qps", 5.0, "QPS to the Kubernetes API server")
	flags.Float64("api-burst", 10, "Burst to the Kubernetes API server")
	return flags
}

func NewControllerFlags() *pflag.FlagSet {
	flags := pflag.NewFlagSet("controller", pflag.ExitOnError)
	flags.String("namespace", "k8s-dra", "Namespace where the controller watches for DeviceAllocation CRDs")
	flags.Int("workers", 3, "Number of workers the controller spawns")
	flags.Int("metrics-port", 9001, "HTTP port to expose metrics")
	flags.String("metrics-path", "metrics", "HTTP path to expose metrics")
	flags.Int("pprof-port", 9002, "HTTP port to expose pprof endpoints")
	return flags
}
