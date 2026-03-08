package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"log/slog"
)

const (
	fooCounterMetric                 = "foo_counter"
	loginSuccessCounterMetric        = "login_success_total"
	loginFailureCounterMetric        = "login_failure_total"
	roleAssignmentCounterMetric      = "role_assignment_total"
	patientCreatedCounterMetric      = "patient_created_total"
	sessionCreatedCounterMetric      = "session_created_total"
	energyModuleCreatedCounterMetric     = "energy_module_created_total"
	injectableModuleCreatedCounterMetric = "injectable_module_created_total"
	outcomeRecordedCounterMetric         = "outcome_recorded_total"
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
		fooCounterMetric:                 newFooCounter(reg),
		loginSuccessCounterMetric:        newLoginSuccessCounter(reg),
		loginFailureCounterMetric:        newLoginFailureCounter(reg),
		roleAssignmentCounterMetric:      newRoleAssignmentCounter(reg),
		patientCreatedCounterMetric:      newPatientCreatedCounter(reg),
		sessionCreatedCounterMetric:      newSessionCreatedCounter(reg),
		energyModuleCreatedCounterMetric:     newEnergyModuleCreatedCounter(reg),
		injectableModuleCreatedCounterMetric: newInjectableModuleCreatedCounter(reg),
		outcomeRecordedCounterMetric:         newOutcomeRecordedCounter(reg),
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

// IncrementRoleAssignmentCount increments the role assignment counter.
func (c *Client) IncrementRoleAssignmentCount() {
	c.metrics[roleAssignmentCounterMetric].(prometheus.Counter).Add(1)
}

// IncrementPatientCreatedCount increments the patient created counter.
func (c *Client) IncrementPatientCreatedCount() {
	c.metrics[patientCreatedCounterMetric].(prometheus.Counter).Add(1)
}

// IncrementSessionCreatedCount increments the session created counter.
func (c *Client) IncrementSessionCreatedCount() {
	c.metrics[sessionCreatedCounterMetric].(prometheus.Counter).Add(1)
}

// IncrementEnergyModuleCreatedCount increments the energy module created counter.
func (c *Client) IncrementEnergyModuleCreatedCount() {
	c.metrics[energyModuleCreatedCounterMetric].(prometheus.Counter).Add(1)
}

// IncrementInjectableModuleCreatedCount increments the injectable module created counter.
func (c *Client) IncrementInjectableModuleCreatedCount() {
	c.metrics[injectableModuleCreatedCounterMetric].(prometheus.Counter).Add(1)
}

// IncrementOutcomeRecordedCount increments the outcome recorded counter.
func (c *Client) IncrementOutcomeRecordedCount() {
	c.metrics[outcomeRecordedCounterMetric].(prometheus.Counter).Add(1)
}
