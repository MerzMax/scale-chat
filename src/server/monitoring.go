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

var MessageProcessingTime = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Namespace: "scale_chat",
		Subsystem: "timing",
		Name:      "processing",
		Help:      "Time to process a message from receiving to sending",
	},
)

func InitMonitoring() {
	prometheus.MustRegister(MessageCounterVec)
	prometheus.MustRegister(MessageProcessingTime)
}
