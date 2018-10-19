package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	remote = "https://billglover-golang.appspot.com/ip"
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

	watchForChange(remote, 10*time.Second)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func watchForChange(url string, d time.Duration) {
	go func() {
		prev := new(net.IP)

		t := time.NewTicker(d)

		for range t.C {
			ip, err := getIP(remote)
			if err != nil {
				log.Println(err)
				continue
			}

			if ip.Equal(*prev) == false {
				changeCount.Inc()
				*prev = ip
			}
		}
	}()
}

func getIP(url string) (net.IP, error) {
	timer := prometheus.NewTimer(reqDurations)
	defer timer.ObserveDuration()

	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to create request")
	}

	client := http.Client{Timeout: 1 * time.Second}
	res, err := client.Do(req)

	if res != nil {
		reqCount.WithLabelValues(strconv.Itoa(res.StatusCode)).Inc()
		defer res.Body.Close()
	}

	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to make request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		reqErrCount.Inc()
		return nil, errors.Wrap(err, "failed to parse response body")
	}
	ip := net.ParseIP(strings.TrimSpace(string(body)))
	return ip, nil
}
