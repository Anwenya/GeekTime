package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"testing"
	"time"
)

func TestGauge(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.GaugeOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_gauge",
		Help:      "测试",
	}

	gauge := prometheus.NewGauge(opts)
	prometheus.MustRegister(gauge)

	for {
		gauge.Inc()
		time.Sleep(time.Second)
		gauge.Dec()
	}
}

func TestGaugeVec(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.GaugeOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_gauge_vec",
		Help:      "测试",
	}

	gauge := prometheus.NewGaugeVec(
		opts,
		[]string{"code", "method"},
	)

	prometheus.MustRegister(gauge)

	gauge.WithLabelValues("404", "POST").Add(120)

	gauge200 := gauge.WithLabelValues("200", "GET")

	for {
		number := rand.Intn(100)

		for i := 0; i < number; i++ {
			gauge200.Inc()
		}
		time.Sleep(time.Second / 2)
		for i := 0; i < number; i++ {
			gauge200.Dec()
		}

	}
}
