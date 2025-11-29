package k8s

import (
	"context"
)

// MetricsClient defines the interface for metrics collection backends
type MetricsClient interface {
	// GetCurrentPodMetrics retrieves current pod metrics from the metrics backend
	GetCurrentPodMetrics(ctx context.Context, namespace string) ([]PodMetric, error)

	// GetHistoricalMetrics retrieves and analyzes 7-day historical metrics for pods
	GetHistoricalMetrics(ctx context.Context, namespace string) ([]HistoricalMetrics, error)

	// GetNamespaces retrieves all namespaces from metrics
	GetNamespaces(ctx context.Context) ([]string, error)

	// Close closes the metrics client connection
	Close() error

	// GetClientType returns the type of metrics client (prometheus, vmagent, etc.)
	GetClientType() string
}

// MetricsClientConfig contains configuration for metrics clients
type MetricsClientConfig struct {
	Backend            string // "prometheus" or "vmagent"
	URL                string // Connection URL for the metrics backend
	Username           string // Optional username for Basic Authentication
	Password           string // Optional password for Basic Authentication
	InsecureSkipVerify bool   // Skip TLS certificate verification (default: false)
	K8sCluster         string // Optional k8s_cluster label filter for VictoriaMetrics
}

// MetricsClientFactory creates metrics clients based on configuration
type MetricsClientFactory struct{}

// NewMetricsClientFactory creates a new metrics client factory
func NewMetricsClientFactory() *MetricsClientFactory {
	return &MetricsClientFactory{}
}

// CreateClient creates a metrics client based on the provided configuration
func (f *MetricsClientFactory) CreateClient(config MetricsClientConfig) (MetricsClient, error) {
	switch config.Backend {
	case "prometheus":
		return NewPrometheusClient(config.URL, config.Username, config.Password, config.InsecureSkipVerify)
	case "victoriametrics":
		return NewVictoriaMetricsClient(config.URL, config.Username, config.Password, config.InsecureSkipVerify, config.K8sCluster)
	default:
		// Default to Prometheus for backward compatibility
		return NewPrometheusClient(config.URL, config.Username, config.Password, config.InsecureSkipVerify)
	}
}
