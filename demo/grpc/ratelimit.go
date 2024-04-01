package grpc

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"sync/atomic"
	"time"
)

// grpc 限流拦截器的不同实现

// CounterLimiter 计数器限流
type CounterLimiter struct {
	count     atomic.Int32
	threshold int32
}

func (l *CounterLimiter) BuilderServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		// 请求进入 计数
		count := l.count.Add(1)
		defer func() {
			l.count.Add(-1)
		}()
		// 通过
		if count <= l.threshold {
			resp, err = handler(ctx, req)
			return
		}
		// 未通过
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// FixedWindowLimiter 固定窗口限流
type FixedWindowLimiter struct {
	// 窗口大小
	window time.Duration
	// 上次窗口起始位置
	lastWindowStart time.Time
	// 窗口内计数
	count int
	// 阈值
	threshold int
	lock      sync.Mutex
}

func (l *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		l.lock.Lock()

		now := time.Now()
		// 已经超过了上一个窗口的范围
		if now.After(l.lastWindowStart.Add(l.window)) {
			l.count = 0
			l.lastWindowStart = now
		}
		// 增加计数
		count := l.count + 1
		l.lock.Unlock()

		// 通过
		if count <= l.threshold {
			resp, err = handler(ctx, req)
			return
		}
		// 未通过
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// SlidingWindowLimiter 滑动窗口限流
type SlidingWindowLimiter struct {
	// 窗口大小
	window time.Duration
	// 请求记录队列
	queue queue.ConcurrentPriorityQueue[time.Time]
	// 阈值
	threshold int
	lock      sync.Mutex
}

func (l *SlidingWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		l.lock.Lock()

		now := time.Now()

		// 快路径
		// 首先看长度是否超过阈值
		// 如果没超过都不用考虑时间
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(now)
			l.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}

		// 已当前请求时间向前推算窗口起始时间
		windowStart := now.Add(-l.window)
		// 删除不在滑动窗口范围内的记录
		for {
			first, _ := l.queue.Peek()
			if first.Before(windowStart) {
				_, _ = l.queue.Dequeue()
			} else {
				break
			}
		}

		// 剩下的都是在本次窗口内的

		// 通过
		if l.queue.Len() < l.threshold {
			_ = l.queue.Enqueue(now)
			l.lock.Unlock()
			resp, err = handler(ctx, req)
			return
		}
		l.lock.Unlock()
		// 未通过
		return nil, status.Errorf(codes.ResourceExhausted, "限流")
	}
}

// TokenBucketLimiter 令牌桶限流
type TokenBucketLimiter struct {
	interval  time.Duration
	buckets   chan struct{}
	closeChan chan struct{}
	closeOnce sync.Once
}

func (l *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	// 启动定时器生成令牌
	ticker := time.NewTicker(l.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				select {
				case l.buckets <- struct{}{}:
				default:
					// 桶满了
				}
			case <-l.closeChan:
				return
			}
		}
	}()

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		select {
		// 通过
		case <-l.buckets:
			return handler(ctx, req)
		// 未通过
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

	}
}

func (l *TokenBucketLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.closeChan)
	})
	return nil
}

// FunnelBucketLimiter 漏斗限流
type FunnelBucketLimiter struct {
	interval  time.Duration
	closeChan chan struct{}
	closeOnce sync.Once
}

func (l *FunnelBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	// 启动定时器生成令牌
	ticker := time.NewTicker(l.interval)
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		select {
		// 通过
		case <-ticker.C:
			return handler(ctx, req)
		// 未通过
		case <-l.closeChan:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		// 未通过
		default:
			return nil, status.Errorf(codes.ResourceExhausted, "限流")
		}

	}
}

func (l *FunnelBucketLimiter) Close() error {
	l.closeOnce.Do(func() {
		close(l.closeChan)
	})
	return nil
}
