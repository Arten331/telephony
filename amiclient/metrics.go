package amiclient

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	messagesReceived *prometheus.CounterVec
	messagesSent     *prometheus.CounterVec
	connectionsTry   *prometheus.CounterVec
}

func newMetrics(service string) *Metrics {
	m := &Metrics{
		messagesReceived: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: service,
				Name:      "ami_messages_received",
			},
			[]string{},
		),
		messagesSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: service,
				Name:      "ami_messages_sent",
			},
			[]string{},
		),
		connectionsTry: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: service,
				Name:      "ami_connections_count",
			},
			[]string{},
		),
	}

	return m
}

func (m *Metrics) getMetrics() []prometheus.Collector {
	collectors := []prometheus.Collector{
		m.messagesSent,
		m.messagesReceived,
		m.connectionsTry,
	}

	return collectors
}

func (m *Metrics) StoreSentMessage() {
	m.messagesSent.WithLabelValues().Inc()
}

func (m *Metrics) StoreReceivedMessage() {
	m.messagesReceived.WithLabelValues().Inc()
}

func (m *Metrics) StoreConnectionCount() {
	m.connectionsTry.WithLabelValues().Inc()
}
