package main

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/metadata"
)

// Prometheus metric : totalGreetings
var totalGreetings = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "learn_grpc_greetings_total",
		Help: "Total number of greetings received by the server",
	},
	[]string{"method", "client_version"},
)

func registerCustomMetrics() {
	prometheus.MustRegister(totalGreetings)
}

func incrementTotalGreetings(ctx context.Context) {
	md, _ := metadata.FromIncomingContext(ctx)
	version := "unknown"
	if v := md.Get(string(RequestVersionKey)); len(v) > 0 {
		version = v[0]
	}

	totalGreetings.WithLabelValues("to_server", version).Inc()
}
