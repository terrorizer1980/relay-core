package opt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nebulav1 "github.com/puppetlabs/relay-core/pkg/apis/nebula.puppet.com/v1"
	"github.com/spf13/viper"
)

var (
	Scheme = runtime.NewScheme()
)

// TODO I thought this was for loading the CRDs in, but nothing calls it now...
func init() {
	builder := runtime.NewSchemeBuilder(
		scheme.AddToScheme,
		nebulav1.AddToScheme,
	)

	if err := builder.AddToScheme(Scheme); err != nil {
		panic(fmt.Sprintf("could not set up scheme for workflow run events: %+v", err))
	}
}

const (
	DefaultProjectID = "nebula-235818"
	DefaultInterval  = 11
)

type Config struct {
	// ProjectID is the GCP project ID to which metrics are logged
	ProjectID string

	// Interval is the number of seconds to wait between each metric submission
	Interval time.Duration

	// Kubeconfig is a path to a kubeconfig file
	Kubeconfig string

	// Environment is the metrics label about which environment is being monitored
	Environment string

	// DeleteMetrics determines if the metric descriptors should be deleted. If descriptors are restored quickly after
	// deletion then the data may still be present if it has not been reaped by GCP yet
	DeleteMetrics bool

	// PublishMetrics determines if the metrics that are collected should be published
	PublishMetrics bool
}

func (c *Config) KubernetesClient() (client.Client, error) {
	// Try homedir config
	config, err := clientcmd.BuildConfigFromFlags("", c.Kubeconfig)
	if err != nil {
		// Try in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	return client.New(config, client.Options{
		Scheme: Scheme,
	})
}

func (c *Config) MetricsClient(ctx context.Context) (*monitoring.MetricClient, error) {
	mc, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}

	return mc, nil

}

func NewConfig() *Config {
	viper.SetDefault("project_id", DefaultProjectID)
	viper.SetDefault("interval", DefaultInterval)
	viper.SetDefault("kubeconfig", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	viper.SetDefault("publish", false)
	viper.SetDefault("environment", "dev")
	viper.SetDefault("delete_descriptiors", false)

	return &Config{
		ProjectID:      viper.GetString("project_id"),
		Interval:       time.Duration(viper.GetInt("interval")),
		Kubeconfig:     viper.GetString("kubeconfig"),
		Environment:    viper.GetString("environment"),
		DeleteMetrics:  viper.GetBool("delete_descriptiors"),
		PublishMetrics: viper.GetBool("publish"),
	}
}
