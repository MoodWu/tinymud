package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// AI调用次数
	AICallTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ai_call_total",
			Help: "Total number of AI calls",
		},
	)

	// AI错误次数
	AIErrorTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ai_error_total",
			Help: "Total number of AI errors",
		},
	)

	// AI延迟（直方图）
	AILatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ai_latency_seconds",
			Help:    "AI call latency",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func Init() {
	prometheus.MustRegister(AICallTotal)
	prometheus.MustRegister(AIErrorTotal)
	prometheus.MustRegister(AILatency)
}
