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
