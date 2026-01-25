package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics
var (
	// Request latency (for response time SLO)
	RequestLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "traffic_steering_request_latency_seconds",
			Help:    "Request latency in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"endpoint", "method"},
	)

	// Request counter (for success rate SLO)
	RequestTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_steering_requests_total",
			Help: "Total number of requests",
		},
		[]string{"endpoint", "method", "status"},
	)

	// Active requests gauge
	ActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traffic_steering_active_requests",
			Help: "Number of currently active requests",
		},
		[]string{"endpoint"},
	)

	// NEF API specific metrics
	NefRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_steering_nef_requests_total",
			Help: "Total NEF API requests",
		},
		[]string{"operation", "status"},
	)

	NefLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "traffic_steering_nef_latency_seconds",
			Help:    "NEF API latency in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"operation"},
	)

	// Prometheus query metrics
	PrometheusQueries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_steering_prometheus_queries_total",
			Help: "Total Prometheus queries",
		},
		[]string{"status"},
	)

	PrometheusLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "traffic_steering_prometheus_latency_seconds",
			Help:    "Prometheus query latency in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{}, // Flattened, no labels in go version for simplicity unless needed
	)

	// Steering operations
	SteeringOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_steering_operations_total",
			Help: "Total steering operations",
		},
		[]string{"target", "status"},
	)

	// Current steering target (info metric)
	CurrentTarget = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "traffic_steering_current_target",
			Help: "Current steering target (1=edge1, 2=edge2, 0=none)",
		},
	)

	// Auto-steering metrics
	AutoSteerTriggers = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "traffic_steering_auto_steer_triggers_total",
			Help: "Auto-steering triggers",
		},
		[]string{"from_target", "to_target", "reason"},
	)

	// UPF traffic rate
	UpfTrafficRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "traffic_steering_upf_traffic_rate_bps",
			Help: "Current traffic rate for each UPF in bytes/sec",
		},
		[]string{"upf"},
	)

	// Auto-steering threshold
	AutoSteerThreshold = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "traffic_steering_auto_steer_threshold_bps",
			Help: "Auto-steering threshold in bytes/sec",
		},
	)
)
