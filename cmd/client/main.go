package main

import (
	"log"
	"net"
	"net/http"

	"github.com/billglover/myIP/pkg/ipmon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reqCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "client",
		Name:      "requests_total",
		Help:      "IP request count",
	}, []string{"code"})
	reqErrCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "client",
		Name:      "requests_err_total",
		Help:      "IP request error count",
	})
	changeCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "client",
		Name:      "changes_total",
		Help:      "the number of times the IP has changed",
	})
	reqDurations = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       "request_durations_seconds",
			Help:       "IP request latency distributions",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
	)
)

func main() {
	prometheus.MustRegister(reqCount, reqErrCount, reqDurations, changeCount)
	http.Handle("/metrics/", promhttp.Handler())

	ch := make(chan net.IP, 1)
	ipmon.WatchExternalIP(ch)
	updateIP(ch)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func updateIP(ips chan net.IP) {
	go func() {
		for {
			select {
			case ip := <-ips:
				log.Println("Update IP:", ip)
			}
		}
	}()
}
