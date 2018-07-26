package main

import (
	"flag"
	"math"
	"strconv"
	"time"
)

var oscillationPeriod = flag.Duration("oscillation-period", 5*time.Minute, "The duration of the rate oscillation period.")

func runClient() {
	oscillationFactor := func() float64 {
		return 2 + math.Sin(math.Sin(2*math.Pi*float64(time.Since(start))/float64(*oscillationPeriod)))
	}

	doGet := func(i int) {
		for {
			handleAPI("GET", "/api/pos-" + strconv.Itoa(i))
			time.Sleep(time.Duration(80*oscillationFactor()) * time.Millisecond)
		}
	}

	for i := 1; i <= routeCount; i++ {
		go doGet(i)
	}

	select {}
}
