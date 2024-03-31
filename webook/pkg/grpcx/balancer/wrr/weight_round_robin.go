package wrr

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync"
)

func init() {
	balancer.Register(newBuilder())
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(
		WeightRoundRobin,
		&WeightedPickerBuilder{},
		base.Config{HealthCheck: true},
	)
}

// 平滑加权轮询

const WeightRoundRobin = "custom_weighted_round_robin"

type weightConn struct {
	// 初始权重
	weight int
	// 当前权重
	currentWeight int
	balancer.SubConn
}

type WeightPicker struct {
	sync.Mutex
	conns []*weightConn
}

func (w *WeightPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// 没有可以的节点
	if len(w.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	var totalWeight int
	var res *weightConn

	w.Lock()
	for _, node := range w.conns {
		totalWeight += node.weight
		node.currentWeight += node.weight
		if res == nil || res.currentWeight < node.currentWeight {
			res = node
		}
	}

	res.currentWeight -= totalWeight
	w.Unlock()
	return balancer.PickResult{
		SubConn: res.SubConn,
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type WeightedPickerBuilder struct{}

func (w *WeightedPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightConn, 0, len(info.ReadySCs))
	for conn, connInfo := range info.ReadySCs {
		weightVal, _ := connInfo.Address.Metadata.(map[string]any)["weight"]
		weight, _ := weightVal.(float64)
		conns = append(conns, &weightConn{
			SubConn:       conn,
			weight:        int(weight),
			currentWeight: int(weight),
		})
	}
	return &WeightPicker{
		conns: conns,
	}
}
