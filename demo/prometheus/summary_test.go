package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"testing"
	"time"
)

func TestSummary(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.SummaryOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_summary",
		Help:      "测试",
		// 0.5-quantile: 0.05意思是允许最后的误差不超过0.05。
		// 假设某个0.5-quantile的值为120，由于设置的误差为0.05，
		// 所以120代表的真实quantile是(0.45, 0.55)范围内的某个值。
		// 之所以要设置误差，原因很简单，就是用一定的误差换取内存空间和CPU计算能力
		Objectives: map[float64]float64{
			0.1:  0.01,
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}

	summary := prometheus.NewSummary(opts)
	prometheus.MustRegister(summary)

	for {
		rand.NewSource(time.Now().UnixNano())
		summary.Observe(rand.Float64())
		time.Sleep(time.Millisecond * 100)
	}
}

func TestSummaryVec(t *testing.T) {
	go StartPrometheusServer()

	opts := prometheus.SummaryOpts{
		Namespace: "prometheus",
		Subsystem: "demo",
		Name:      "test_summary_vec",
		Help:      "测试",
		Objectives: map[float64]float64{
			0.1:  0.01,
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001,
		},
	}

	summary := prometheus.NewSummaryVec(
		opts,
		[]string{"code", "method"},
	)

	prometheus.MustRegister(summary)

	summary.WithLabelValues("404", "POST").Observe(5)

	summary200 := summary.WithLabelValues("200", "GET")

	for {
		rand.NewSource(time.Now().UnixNano())
		summary200.Observe(rand.Float64())
		time.Sleep(time.Millisecond * 100)
	}
}
