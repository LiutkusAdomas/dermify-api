package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"log/slog"
)

const (
	fooCounterMetric          = "foo_counter"
	loginSuccessCounterMetric = "login_success_total"
	loginFailureCounterMetric = "login_failure_total"
)

// Client allows the creation and invocation of metrics within the API. Instantiation should occur through the New
// function as it creates internal resources.
type Client struct {
	logger   *slog.Logger
	metrics  map[string]prometheus.Metric
	Registry *prometheus.Registry
}

// New is the intended way to instantiate an metrics Client. This method should be used over direct instantiation
// because it creates internal resources.
func New(logger *slog.Logger) *Client {
	reg := prometheus.NewRegistry()

	metrics := map[string]prometheus.Metric{
		fooCounterMetric:          newFooCounter(reg),
		loginSuccessCounterMetric: newLoginSuccessCounter(reg),
		loginFailureCounterMetric: newLoginFailureCounter(reg),
	}

	return &Client{
		logger:   logger,
		metrics:  metrics,
		Registry: reg,
	}
}

func (c *Client) IncrementFooCount() {
	c.metrics[fooCounterMetric].(prometheus.Counter).Add(1)
}

func (c *Client) IncrementLoginSuccessCount() {
	c.metrics[loginSuccessCounterMetric].(prometheus.Counter).Add(1)
}

func (c *Client) IncrementLoginFailureCount() {
	c.metrics[loginFailureCounterMetric].(prometheus.Counter).Add(1)
}
