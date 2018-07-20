package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}

var (
	endpointCount, err = strconv.Atoi(getenv("ENDPOINT_COUNT", "5"))
	routeCount, _ = strconv.Atoi(getenv("ROUTE_COUNT", "5"))
	n = flag.Int(
		"endpoint-count",  endpointCount,
		"Number of sequential endpoints to serve metrics on, starting at /metrics/1",
	)
	registerProcessMetrics = flag.Bool(
		"enable-process-metrics", true,
		"Include (potentially expensive) process_* metrics.",
	)
	registerGoMetrics = flag.Bool(
		"enable-go-metrics", true,
		"Include (potentially expensive) go_* metrics.",
	)
	allowCompression = flag.Bool(
		"allow-metrics-compression", true,
		"Allow gzip compression of metrics.",
	)

	start = time.Now()
)

func main() {
	log.Print("Running " + strconv.Itoa(endpointCount) + " endpoints with " + strconv.Itoa(routeCount) + " routes")

	flag.Parse()

	if *registerProcessMetrics {
		registry.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
	}
	if *registerGoMetrics {
		registry.MustRegister(prometheus.NewGoCollector())
	}

	mux := http.NewServeMux()
	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fmt.Sprintf("Observe test metrics generator: %d endpoints, %d routes per endpoint", endpointCount, routeCount))
	})

	for i := 1; i <= *n; i++ {
		mux.Handle("/metrics/" + strconv.Itoa(i), promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{
				DisableCompression: !*allowCompression,
			},
		))
	}

	go http.ListenAndServe(":8080", mux)

	runClient()
}
