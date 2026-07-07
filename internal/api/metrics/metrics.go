package metrics

import "github.com/prometheus/client_golang/prometheus"

// PromQL: rate(custos_http_requests_total[5m])
var HTTPRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "custos",
		Name:      "http_requests_total",
		Help:      "Total number of HTTP requests by method, route, and status.",
	},
	[]string{"method", "route", "status"},
)

// PromQL: histogram_quantile(0.99, sum(rate(custos_http_duration_seconds_bucket[5m])) by (le, route))
var HTTPDurationSeconds = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "custos",
		Name:      "http_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"method", "route"},
)

// PromQL: rate(custos_events_ingested_total[5m])
var EventsIngestedTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "custos",
	Name:      "events_ingested_total",
	Help:      "Total number of raw events received by the ingest endpoint.",
})

// PromQL: rate(custos_issues_created_total[5m])
var IssuesCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "custos",
	Name:      "issues_created_total",
	Help:      "Total number of new issues opened (fingerprint not seen before).",
})

// PromQL: rate(custos_analysis_total[5m])
var AnalysisTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "custos",
		Name:      "analysis_total",
		Help:      "Total AI analysis results by outcome (success, failed).",
	},
	[]string{"outcome"},
)

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPDurationSeconds,
		EventsIngestedTotal,
		IssuesCreatedTotal,
		AnalysisTotal,
	)
}
