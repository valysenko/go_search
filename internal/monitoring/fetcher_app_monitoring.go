package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type FetcherPrometheusMetricsService struct {
	registry           *prometheus.Registry
	runArticlesTotal   *prometheus.CounterVec
	runErrorsTotal     *prometheus.CounterVec
	runDurationSeconds *prometheus.HistogramVec
	pushgatewayURL     string
	jobName            string
}

// TODO: figh cardinality because of run_id label
func NewFetcherPrometheusMetricsService(namespace, subsystem, podName string, pushgatewayURL string) *FetcherPrometheusMetricsService {
	fetcherJobName := "fetcher_job"
	constLabels := map[string]string{"podName": podName}
	registry := prometheus.NewRegistry()

	runArticlesTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			ConstLabels: constLabels,
			Name:        "run_articles_total",
			Help:        "Total number of articles fetched.",
		},
		[]string{"provider", "category", "run_id"},
	)

	runErrorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			ConstLabels: constLabels,
			Name:        "run_errors_total",
			Help:        "Total number of run errors.",
		},
		[]string{"category", "run_id"},
	)

	runDurationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			ConstLabels: constLabels,
			Name:        "run_duration_seconds",
			Help:        "Histogram of run durations.",
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"run_id"},
	)

	p := &FetcherPrometheusMetricsService{
		registry:           registry,
		runArticlesTotal:   runArticlesTotal,
		runErrorsTotal:     runErrorsTotal,
		runDurationSeconds: runDurationSeconds,
		pushgatewayURL:     pushgatewayURL,
		jobName:            fetcherJobName,
	}

	p.registerMetrics()

	return p
}

func (p *FetcherPrometheusMetricsService) registerMetrics() {
	p.registry.MustRegister(
		p.runArticlesTotal, p.runErrorsTotal, p.runDurationSeconds,
	)
}

func (p *FetcherPrometheusMetricsService) Registry() *prometheus.Registry {
	return p.registry
}

func (p *FetcherPrometheusMetricsService) IncrementRunArticlesTotal(provider, category, runID string) {
	p.runArticlesTotal.WithLabelValues(provider, category, runID).Inc()
}

func (p *FetcherPrometheusMetricsService) IncrementErrorsTotal(category, runID string) {
	p.runErrorsTotal.WithLabelValues(category, runID).Inc()
}

func (p *FetcherPrometheusMetricsService) ObserveRunDuration(duration time.Duration, runID string) {
	durationSeconds := duration.Seconds()
	p.runDurationSeconds.WithLabelValues(runID).Observe(durationSeconds)
}

func (p *FetcherPrometheusMetricsService) Push() error {
	return push.New(p.pushgatewayURL, p.jobName).
		Gatherer(p.registry).
		Push()
}
