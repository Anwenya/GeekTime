package scheduler

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	"github.com/Anwenya/GeekTime/webook/pkg/gormx/connpool"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"sync"
	"time"
)

// Scheduler 用来统一管理整个迁移过程
type Scheduler[T migrator.Entity] struct {
	lock       sync.Mutex
	src        *gorm.DB
	dst        *gorm.DB
	pool       *connpool.DoubleWritePool
	l          logger.LoggerV1
	pattern    string
	cancelFull func()
	cancelIncr func()
	producer   events.Producer

	// 如果你要允许多个全量校验同时运行
	fulls map[string]func()
}

func NewScheduler[T migrator.Entity](
	l logger.LoggerV1,
	src *gorm.DB,
	dst *gorm.DB,
	// 这个是业务用的 DoubleWritePool
	pool *connpool.DoubleWritePool,
	producer events.Producer,
) *Scheduler[T] {
	return &Scheduler[T]{
		src:        src,
		dst:        dst,
		pattern:    connpool.PatternSrcOnly,
		cancelFull: func() {},
		cancelIncr: func() {},
		pool:       pool,
		producer:   producer,
		l:          l,
	}
}

func (s *Scheduler[T]) RegisterRoutes(server *gin.RouterGroup) {
	server.POST("/src_only", decorator.Wrap(s.SrcOnly))
	server.POST("/src_first", decorator.Wrap(s.SrcFirst))
	server.POST("/dst_first", decorator.Wrap(s.DstFirst))
	server.POST("/dst_only", decorator.Wrap(s.DstOnly))
	server.POST("/full/start", decorator.Wrap(s.StartFullValidation))
	server.POST("/full/stop", decorator.Wrap(s.StopFullValidation))
	server.POST("/incr/stop", decorator.Wrap(s.StopIncrementValidation))
	server.POST("/incr/start", decorator.WrapBody[StartIncrRequest](s.StartIncrementValidation))
}

// ---- 下面是四个阶段 ---- //

// SrcOnly 只读写源表
func (s *Scheduler[T]) SrcOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcOnly
	s.pool.UpdatePattern(connpool.PatternSrcOnly)
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) SrcFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternSrcFirst
	s.pool.UpdatePattern(connpool.PatternSrcFirst)
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) DstFirst(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstFirst
	s.pool.UpdatePattern(connpool.PatternDstFirst)
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) DstOnly(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.pattern = connpool.PatternDstOnly
	s.pool.UpdatePattern(connpool.PatternDstOnly)
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) StopIncrementValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelIncr()
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) StartIncrementValidation(c *gin.Context, req StartIncrRequest) (ginx.Result, error) {
	// 开启增量校验
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelIncr
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, nil
	}
	v.Incr().UpdateTime(req.UpdateTime).SleepInterval(time.Duration(req.Interval) * time.Millisecond)

	go func() {
		var ctx context.Context
		ctx, s.cancelIncr = context.WithCancel(context.Background())
		cancel()
		err := v.Validate(ctx)
		s.l.Warn("退出增量校验", logger.Error(err))
	}()
	return ginx.Result{Msg: "启动增量校验成功"}, nil
}

func (s *Scheduler[T]) StopFullValidation(c *gin.Context) (ginx.Result, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelFull()
	return ginx.Result{Msg: "OK"}, nil
}

// StartFullValidation 全量校验
func (s *Scheduler[T]) StartFullValidation(c *gin.Context) (ginx.Result, error) {
	// 可以考虑去重的问题
	s.lock.Lock()
	defer s.lock.Unlock()
	// 取消上一次的
	cancel := s.cancelFull
	v, err := s.newValidator()
	if err != nil {
		return ginx.Result{}, err
	}
	var ctx context.Context
	ctx, s.cancelFull = context.WithCancel(context.Background())

	go func() {
		// 先取消上一次的
		cancel()
		err := v.Validate(ctx)
		if err != nil {
			s.l.Warn("退出全量校验", logger.Error(err))
		}
	}()
	return ginx.Result{Msg: "OK"}, nil
}

func (s *Scheduler[T]) newValidator() (*validator.Validator[T], error) {
	switch s.pattern {
	case connpool.PatternSrcOnly, connpool.PatternSrcFirst:
		return validator.NewValidator[T](s.src, s.dst, "SRC", s.l, s.producer), nil
	case connpool.PatternDstFirst, connpool.PatternDstOnly:
		return validator.NewValidator[T](s.dst, s.src, "DST", s.l, s.producer), nil
	default:
		return nil, fmt.Errorf("未知的 pattern %s", s.pattern)
	}
}

type StartIncrRequest struct {
	UpdateTime int64 `json:"update_time"`
	// 毫秒数
	// json 不能正确处理 time.Duration 类型
	Interval int64 `json:"interval"`
}
