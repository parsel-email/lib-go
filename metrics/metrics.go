// Package metrics provides Prometheus metrics for the auth service
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Requests tracks authentication requests by provider and status
	Requests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "The total number of authentication requests",
		},
		[]string{"provider", "status"},
	)

	// TokenOperations tracks token operations by type (issue, refresh, revoke) and status
	TokenOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_operations_total",
			Help: "The total number of token operations",
		},
		[]string{"operation", "status"},
	)

	// APIKeyOperations tracks API key operations by type (create, list, revoke) and status
	APIKeyOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "apikey_operations_total",
			Help: "The total number of API key operations",
		},
		[]string{"operation", "status"},
	)

	// RequestDuration tracks API request durations
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "The duration of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	// DatabaseOperations tracks database operations and their status
	DatabaseOperations = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_operations_total",
			Help: "The total number of database operations",
		},
		[]string{"operation", "status"},
	)

	// ActiveSessions tracks the number of active user sessions
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "The number of active user sessions",
		},
	)

	// Errors tracks error occurrences by type
	Errors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "The total number of errors",
		},
		[]string{"type"},
	)
)
