package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"testing"
	"time"
)

func TestHistogram(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.HistogramOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_histogram",
		Help:      "测试",
		// 分成 9个区域 <=1 , >1&&<=2 ... >8&&<=9  最后超过9的部分会被记录到inf中
		Buckets: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	histogram := prometheus.NewHistogram(opts)
	prometheus.MustRegister(histogram)

	for {
		histogram.Observe(float64(rand.Intn(10)))
		time.Sleep(time.Millisecond * 100)
	}
}

func TestHistogramVec(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.HistogramOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_histogram_vec",
		Help:      "测试",
		// 分成 9个区域 <=1 , >1&&<=2 ... >8&&<=9  最后超过9的部分会被记录到inf中
		Buckets: []float64{1, 2, 3, 4, 5, 6, 7, 8, 9},
	}

	histogram := prometheus.NewHistogramVec(
		opts,
		[]string{"code", "method"},
	)

	prometheus.MustRegister(histogram)

	histogram.WithLabelValues("404", "POST").Observe(5)

	histogram200 := histogram.WithLabelValues("200", "GET")

	for {
		histogram200.Observe(float64(rand.Intn(10)))
		time.Sleep(time.Millisecond * 100)

	}
}
