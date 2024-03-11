package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"testing"
	"time"
)

func StartPrometheusServer() {
	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(any(err))
	}
}

func TestCounter(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_counter",
		Help:      "测试",
	}

	counter := prometheus.NewCounter(opts)
	prometheus.MustRegister(counter)

	for {
		counter.Inc()
		time.Sleep(time.Second)
	}
}

func TestCounterVec(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.CounterOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_counter_vec",
		Help:      "测试",
	}

	counter := prometheus.NewCounterVec(
		opts,
		[]string{"code", "method"},
	)
	prometheus.MustRegister(counter)

	counter.WithLabelValues("404", "POST").Add(42)

	counter200 := counter.WithLabelValues("200", "GET")

	for {
		counter200.Inc()
		time.Sleep(time.Second)
	}
}
