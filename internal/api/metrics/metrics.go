package metrics

import "github.com/prometheus/client_golang/prometheus"

func newFooCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "foo_request_total",
		Help: "Total number of requests to the foo endpoint",
	})
	reg.MustRegister(m)
	return m
}

func newLoginSuccessCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "login_success_total",
		Help: "Total number of successful login attempts",
	})
	reg.MustRegister(m)
	return m
}

func newLoginFailureCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "login_failure_total",
		Help: "Total number of failed login attempts",
	})
	reg.MustRegister(m)
	return m
}

func newRoleAssignmentCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "role_assignment_total",
		Help: "Total number of role assignments.",
	})
	reg.MustRegister(m)
	return m
}

func newPatientCreatedCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "patient_created_total",
		Help: "Total number of patients created.",
	})
	reg.MustRegister(m)
	return m
}

func newSessionCreatedCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "session_created_total",
		Help: "Total number of sessions created.",
	})
	reg.MustRegister(m)
	return m
}

func newEnergyModuleCreatedCounter(reg *prometheus.Registry) prometheus.Counter {
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "dermify_energy_module_created_total",
		Help: "Total number of energy modules created.",
	})
	reg.MustRegister(m)
	return m
}
