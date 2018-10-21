package ipmon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	reqCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "ipmon",
		Name:      "GetExternalIP_req_total",
		Help:      "IP request count",
	})
	reqErrCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "ipmon",
		Name:      "GetExternalIP_err_total",
		Help:      "IP request error count",
	})
	reqDurations = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "myip",
		Subsystem:  "ipmon",
		Name:       "GetExternalIP_request_durations_seconds",
		Help:       "IP request latency distributions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"code"})
)
