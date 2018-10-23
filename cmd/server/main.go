package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	resCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "server",
		Name:      "response_count_total",
		Help:      "Response count",
	}, []string{"code"})
	resDurations = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:  "myip",
		Subsystem:  "server",
		Name:       "response_durations_seconds",
		Help:       "Response latency distributions",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"code"})
)

func main() {
	http.HandleFunc("/ip",promhttp.InstrumentHandlerCounter(resCount, promhttp.InstrumentHandlerDuration(resDurations, http.HandlerFunc(ipHandler))))
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics/", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("environment variable PORT is not set")
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func ipHandler(res http.ResponseWriter, req *http.Request) {
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Fprintln(res, ip)
}
