package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	remote = "https://billglover-golang.appspot.com/ip"
)

var (
	requestCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "client",
		Name: "request_count",
		Help: "the number of requests to the IP lookup URL",
	})
	changeCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "myip",
		Subsystem: "client",
		Name: "change_count",
		Help: "the number of times the IP has changed",
	})
)

func main() {
	prometheus.MustRegister(requestCount, changeCount)

	ips := make(chan net.IP, 1)
	updates := make(chan net.IP, 1)
	go lookupIP(remote, 10 * time.Second, ips)
	go watchForChange(ips, updates)

	http.Handle("/metrics/", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func lookupIP(url string, d time.Duration, ips chan<- net.IP) {
	t := time.NewTicker(d)

	for range t.C {
		ip, err := getIP(remote)
		if err != nil {
			log.Println(err)
			continue
		}
		ips <- ip
	}
}

func getIP(url string) (net.IP, error) {
	// TODO:
	// - must time out quickly
	// - must keep metrics on failures
	// - must keep metrics on response times
	requestCount.Inc()

	req, err := http.NewRequest(http.MethodGet, remote, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}

	client := http.DefaultClient
	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to make request")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse response body")
	}
	ip := net.ParseIP(strings.TrimSpace(string(body)))
	return ip, nil
}

func watchForChange(ips <-chan net.IP, updates chan <-net.IP) {
	prev := new(net.IP)
	for ip := range ips {
		fmt.Println("current:", ip, "prev:", prev)
		if ip.Equal(*prev) == false {
			changeCount.Inc()
			*prev = ip
			updates<-ip
		}
	}
}
