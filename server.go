package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	registry = prometheus.NewRegistry()

	namespace = "observe_loadtest"
	subsystem = "api"

	requestHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "A histogram of the API HTTP request durations in seconds.",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 1.5, 25),
		},
		[]string{"method", "path", "status"},
	)
	requestsInProgress = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "http_requests_in_progress",
			Help:      "The current number of API HTTP requests in progress.",
		})
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of requests",
		},
		[]string{"method", "path", "status"},
	)
	requestErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_errors_total",
			Help:      "Total number of request errors",
		},
		[]string{"method", "path", "status"},
	)
	testCounterTotalInit = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "test_init_counter_total",
			Help:      "Total number of init counter",
		},
		[]string{"test_label"},
	)
	testCounterTotalLaterInit = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "test_later_init_counter_total",
			Help:      "Total number of later init counter",
		},
		[]string{"test_label"},
	)
	testCounterTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "test_counter_total",
			Help:      "Total number of counter",
		},
		[]string{"test_label"},
	)
)

func init() {
	registry.MustRegister(
		requestsTotal,
		requestErrorsTotal,
		requestHistogram,
		requestsInProgress,
		testCounterTotalInit,
		testCounterTotalLaterInit,
	)

	testCounterTotalInit.WithLabelValues("test_init")
}

type responseOpts struct {
	baseLatency time.Duration
	errorRatio  float64

	// Whenever 10*outageDuration has passed, an outage will be simulated
	// that lasts for outageDuration. During the outage, errorRatio is
	// increased by a factor of 10, and baseLatency by a factor of 3.  At
	// start-up time, an outage is simulated, too (so that you can see the
	// effects right ahead and don't have to wait for 10*outageDuration).
	outageDuration time.Duration
}


func handleAPI(method, path string) {
	getRandomResponseOpts := func() responseOpts {
		randFloat := func(min, max float64) float64 {
			return min + rand.Float64() * (max - min)
		}

		return responseOpts{
			baseLatency:    time.Duration(1 + rand.Intn(100)) * time.Millisecond,
			errorRatio:     randFloat(0.001, 0.05),
			outageDuration: time.Duration(1 + rand.Intn(50)) * time.Second,
		}
	}

	requestsInProgress.Inc()
	status := http.StatusOK
	duration := time.Millisecond

	defer func() {
		requestsInProgress.Dec()
		requestHistogram.With(prometheus.Labels{
			"method": method,
			"path":   path,
			"status": fmt.Sprint(status),
		}).Observe(duration.Seconds())
		requestsTotal.WithLabelValues(method, path, fmt.Sprint(status)).Inc()
	}()

	response := getRandomResponseOpts()

	latencyFactor := time.Duration(1)
	errorFactor := 1.
	if time.Since(start)%(10*response.outageDuration) < response.outageDuration {
		latencyFactor *= 3
		errorFactor *= 10
	}
	duration = (response.baseLatency + time.Duration(rand.NormFloat64()*float64(response.baseLatency)/10)) * latencyFactor

	if rand.Float64() <= response.errorRatio*errorFactor {
		status = http.StatusInternalServerError
		requestErrorsTotal.WithLabelValues(method, path, fmt.Sprint(status)).Inc()
	}
}

func handleTestAPI(testLabel string, register bool) {
	log.Print("Test counter")
	if register {
		registry.MustRegister(
			testCounterTotal,
		)
		testCounterTotalLaterInit.WithLabelValues("test_later_init")
	}
	testCounterTotal.WithLabelValues(testLabel).Inc()
	testCounterTotalInit.WithLabelValues("test_init").Inc()
	testCounterTotalLaterInit.WithLabelValues("test_later_init").Inc()
}
