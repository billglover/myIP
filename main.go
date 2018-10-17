package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/ip", ipHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func ipHandler(res http.ResponseWriter, req *http.Request) {

	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		log.Println("/ip, 404, X-Forwarded-For not set")
		res.WriteHeader(http.StatusNotFound)
	}
	log.Printf("/ip, 200, %s\n", ip)
	fmt.Fprintln(res, ip)
}
