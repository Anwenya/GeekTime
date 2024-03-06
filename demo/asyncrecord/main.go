package main

import (
	"fmt"
	"math"
	"time"
)

type record struct {

	// 连续超时次数
	ctt int
	// 超时阈值
	tto time.Duration
	// 最大连续超时次数
	mctt int

	// 记录响应时间
	rtl []int64
	// 参与计算平均响应时间的次数
	aq int
	// 平均响应时间阈值
	tat time.Duration
	// 近几次总响应时间
	srt int64

	// 最大/小响应时间
	maxrt int64
	minrt int64
}

func newRecord(tto time.Duration, mctt int, aq int, tat time.Duration) *record {
	return &record{
		tto:   tto,
		mctt:  mctt,
		aq:    aq,
		tat:   tat,
		rtl:   make([]int64, aq, aq),
		maxrt: math.MinInt,
		minrt: math.MaxInt,
	}
}

func (r *record) Record(rt time.Duration) {
	rtInt64 := rt.Milliseconds()

	// 计算总响应时间
	r.srt += rtInt64
	r.srt -= r.rtl[0]

	// 记录近几次的响应时间 从后往前添加
	copy(r.rtl, r.rtl[1:r.aq])
	r.rtl[r.aq-1] = rtInt64

	// 记录连续超时次数
	if rt > r.tto {
		r.ctt += 1
	} else {
		r.ctt = 0
	}

	// 最大与最小响应时间
	if rtInt64 > r.maxrt {
		r.maxrt = rtInt64
	}
	if rtInt64 < r.minrt {
		r.minrt = rtInt64
	}
}

func (r *record) Judge() {
	fmt.Println(r.srt)
	fmt.Println(r.rtl)
	fmt.Println(r.ctt)
	fmt.Println(r.maxrt)
	fmt.Println(r.minrt)
}

func main() {
	r := newRecord(time.Second*7, 5, 10, time.Second*5)

	for i := 1; i < 15; i++ {

		r.Record(time.Duration(1000000000 * i))
		r.Judge()
	}

	s := make([]int64, 0, 6)
	fmt.Println(len(s))
}
