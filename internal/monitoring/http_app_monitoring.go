package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type PrometheusMetricsService struct {
	registry             *prometheus.Registry
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestsDuration *prometheus.HistogramVec
}

func NewPrometheusMetricsService(namespace, subsystem, podName string) *PrometheusMetricsService {
	constLabels := map[string]string{"podName": podName}
	registry := prometheus.NewRegistry()

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			ConstLabels: constLabels,
			Name:        "requests_total",
			Help:        "Total number of HTTP requests received.",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			ConstLabels: constLabels,
			Name:        "request_duration_seconds",
			Help:        "Histogram of HTTP request durations.",
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status_code"},
	)

	p := &PrometheusMetricsService{
		registry:             registry,
		httpRequestsTotal:    httpRequestsTotal,
		httpRequestsDuration: httpRequestDuration,
	}

	p.registerMetrics()

	return p
}

func (p *PrometheusMetricsService) registerMetrics() {
	p.registry.MustRegister(
		p.httpRequestsTotal, p.httpRequestsDuration,
	)

	p.registry.MustRegister(
		collectors.NewGoCollector(),                                       // Goroutines, GC stats, memory stats (go_goroutines, go_gc_*, go_memstats_*)
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}), // OS-level process metrics — CPU, RSS, open FDs (process_cpu_seconds_total, process_resident_memory_bytes, etc.)
		collectors.NewBuildInfoCollector(),                                // Single go_build_info metric with Go version, module path as labels
	)
}

func (p *PrometheusMetricsService) Registry() *prometheus.Registry {
	return p.registry
}

func (p *PrometheusMetricsService) IncrementRequestsTotal(method, endpoint, statusCode string) {
	p.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
}

func (p *PrometheusMetricsService) ObserveRequestDuration(method, endpoint, statusCode string, durationSeconds float64) {
	p.httpRequestsDuration.WithLabelValues(method, endpoint, statusCode).Observe(durationSeconds)
}
