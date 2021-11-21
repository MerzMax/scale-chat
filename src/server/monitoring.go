package main

import "github.com/prometheus/client_golang/prometheus"

var MessageCounterVec = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "scale_chat",
		Subsystem: "messages",
		Name:      "total",
		Help:      "Total number of processed messages",
	},
	[]string{"type"},
)

func InitMonitoring() {
	prometheus.MustRegister(MessageCounterVec)
}