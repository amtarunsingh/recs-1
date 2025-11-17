package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics collectors
type Metrics struct {
	// HTTP Request metrics
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsTotal   *prometheus.CounterVec

	// Vote operation metrics
	VotesTotal        *prometheus.CounterVec
	VoteErrorsTotal   *prometheus.CounterVec
	VoteChangesTotal  *prometheus.CounterVec
	VoteDeletionTotal *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(registry *prometheus.Registry) *Metrics {
	factory := promauto.With(registry)

	return &Metrics{
		// HTTP Request duration histogram with buckets optimized for API responses
		HTTPRequestDuration: factory.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			},
			[]string{"method", "path", "status"},
		),

		// Total HTTP requests counter
		HTTPRequestsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		// Total votes counter by vote type
		VotesTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "votes_total",
				Help: "Total number of votes processed by type",
			},
			[]string{"vote_type", "operation"},
		),

		// Vote errors counter by error type
		VoteErrorsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vote_errors_total",
				Help: "Total number of vote operation errors by type",
			},
			[]string{"operation", "error_type"},
		),

		// Vote changes counter by vote type transition
		VoteChangesTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vote_changes_total",
				Help: "Total number of vote changes by type transition",
			},
			[]string{"from_type", "to_type"},
		),

		// Vote deletions counter
		VoteDeletionTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vote_deletions_total",
				Help: "Total number of vote deletions",
			},
			[]string{"vote_type"},
		),
	}
}

// RecordHTTPRequest records an HTTP request with duration
func (m *Metrics) RecordHTTPRequest(method, path, status string, duration float64) {
	m.HTTPRequestDuration.WithLabelValues(method, path, status).Observe(duration)
	m.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
}

// RecordVoteAdded records a vote addition
func (m *Metrics) RecordVoteAdded(voteType string) {
	m.VotesTotal.WithLabelValues(voteType, "add").Inc()
}

// RecordVoteChanged records a vote change
func (m *Metrics) RecordVoteChanged(fromType, toType string) {
	m.VoteChangesTotal.WithLabelValues(fromType, toType).Inc()
	m.VotesTotal.WithLabelValues(toType, "change").Inc()
}

// RecordVoteDeleted records a vote deletion
func (m *Metrics) RecordVoteDeleted(voteType string) {
	m.VoteDeletionTotal.WithLabelValues(voteType).Inc()
	m.VotesTotal.WithLabelValues(voteType, "delete").Inc()
}

// RecordVoteError records a vote operation error
func (m *Metrics) RecordVoteError(operation, errorType string) {
	m.VoteErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
