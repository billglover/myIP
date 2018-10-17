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

var ipsRequested = promauto.NewCounter(prometheus.CounterOpts{
	Name: "myip_requests_total",
	Help: "The total number of IPs requested",
})

func main() {
	http.HandleFunc("/ip", ipHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics/", promhttp.Handler())
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func ipHandler(res http.ResponseWriter, req *http.Request) {

	ipsRequested.Inc()

	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		log.Println("/ip, 404, X-Forwarded-For not set")
		res.WriteHeader(http.StatusNotFound)
		return
	}
	log.Printf("/ip, 200, %s\n", ip)
	fmt.Fprintln(res, ip)
}
