package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/billglover/myIP/pkg/ipmon"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	changeCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "ipmon",
		Name:      "UpdateIP_req_total",
		Help:      "IP update request count",
	})
	changeErrCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "ipmon",
		Name:      "UpdateIP_err_total",
		Help:      "IP update error count",
	})
)

func main() {
	http.Handle("/metrics/", promhttp.Handler())

	ch := make(chan net.IP, 1)
	ipmon.WatchExternalIP(ch, 10*time.Second)
	updateIP(ch)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func updateIP(ips chan net.IP) {
	go func() {
		for {
			select {
			case ip := <-ips:
				changeCount.Inc()
				log.Println("Update IP:", ip)
			}
		}
	}()
}
