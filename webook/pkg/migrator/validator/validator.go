package validator

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"time"
)

type Validator[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	producer  events.Producer
	direction string
	batchSize int

	updateTime int64
	// <= 0 就认为中断
	// > 0 就认为睡眠
	sleepInterval time.Duration
	fromBase      func(ctx context.Context, offset int) ([]T, error)

	l logger.LoggerV1
}

func NewValidator[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	producer events.Producer,
	direction string,
	l logger.LoggerV1,
) *Validator[T] {

	validator := &Validator[T]{
		base:      base,
		target:    target,
		producer:  producer,
		direction: direction,
		l:         l,

		batchSize: 100,
	}
	validator.fromBase = validator.fullFromBase
	return validator
}

func (v *Validator[T]) Validate(ctx context.Context) error {
	var eg errgroup.Group
	eg.Go(func() error {
		return v.validateBase2Target(ctx)
	})

	eg.Go(func() error {
		return v.validateTarget2Base(ctx)
	})

	return eg.Wait()
}

// 以 base 为准
func (v *Validator[T]) validateBase2Target(ctx context.Context) error {
	offset := 0
	for {
		// 一次查询出一批 然后逐个校验
		bs, err := v.fromBase(ctx, offset)

		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}

		// 无结果 结束
		if err == gorm.ErrRecordNotFound || len(bs) == 0 {
			// 增量校验 要考虑一直运行
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
			continue
		}

		// 失败跳过
		if err != nil {
			v.l.Error(
				"base => target 查询 base",
				logger.Error(err),
			)
			offset += len(bs)
			continue
		}

		// 查询 target 与 上面 base 的结果进行比较
		var dstTs []T
		// 取出id
		ids := slice.Map[T](
			bs,
			func(idx int, t T) int64 {
				return t.ID()
			},
		)
		err = v.target.WithContext(ctx).
			Where("id IN ?", ids).
			Find(&dstTs).Error

		// 目标表没有数据
		if err == gorm.ErrRecordNotFound || len(dstTs) == 0 {
			// 生成对应的修复任务
			v.notifyTargetMissing(bs)
			offset += len(bs)
			continue
		}

		// 其他异常
		if err != nil {
			v.l.Error(
				"target => base 查询 base 失败",
				logger.Error(err),
			)

			// 保守起见 我都认为 target 里面没有数据
			// v.notifyTargetMissing(bs)
			offset += len(bs)
			continue
		}

		// 找差集 diff代表 在 base 存在 target 不存在
		diff := slice.DiffSetFunc[T](
			bs,
			dstTs,
			func(src, dst T) bool {
				return src.CompareTo(dst)
			},
		)
		v.notifyTargetMissing(diff)

		// 结束
		if len(bs) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(bs)
	}
}

// 以 target 为准
func (v *Validator[T]) validateTarget2Base(ctx context.Context) error {
	offset := 0
	for {
		var ts []T
		// 一次查询出一批 然后逐个校验
		err := v.target.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error

		if err == context.DeadlineExceeded || err == context.Canceled {
			return nil
		}

		// 无结果 结束或者继续
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			if v.sleepInterval <= 0 {
				return nil
			}

			time.Sleep(v.sleepInterval)
			continue
		}

		// 失败跳过
		if err != nil {
			v.l.Error(
				"target => base 查询 target失败",
				logger.Error(err),
			)
			offset += len(ts)
			continue
		}

		// 查询 base 与 上面target的结果进行比较
		var srcTs []T
		// 取出id
		ids := slice.Map[T](
			ts,
			func(idx int, t T) int64 {
				return t.ID()
			},
		)
		err = v.base.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&srcTs).Error

		// 源表没有数据
		if err == gorm.ErrRecordNotFound || len(srcTs) == 0 {
			// 生成对应的修复任务
			v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}

		// 其他异常
		if err != nil {
			v.l.Error(
				"target => base 查询 base 失败",
				logger.Error(err),
			)

			// 保守起见 我都认为 base 里面没有数据
			// v.notifyBaseMissing(ts)
			offset += len(ts)
			continue
		}

		// 找差集 diff代表 在target存在 base不存在
		diff := slice.DiffSetFunc[T](
			ts,
			srcTs,
			func(src, dst T) bool {
				return src.ID() == dst.ID()
			},
		)
		v.notifyBaseMissing(diff)

		// 结束
		if len(ts) < v.batchSize {
			if v.sleepInterval <= 0 {
				return nil
			}
			time.Sleep(v.sleepInterval)
		}
		offset += len(ts)
	}
}

func (v *Validator[T]) notifyBaseMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeBaseMissing)
	}
}

func (v *Validator[T]) notifyTargetMissing(ts []T) {
	for _, val := range ts {
		v.notify(val.ID(), events.InconsistentEventTypeTargetMissing)
	}
}

func (v *Validator[T]) notify(id int64, typ string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 生成对应校验方向的修复任务
	err := v.producer.ProduceInconsistentEvent(
		ctx,
		events.InconsistentEvent{
			ID:        id,
			Type:      typ,
			Direction: v.direction,
		},
	)
	if err != nil {
		v.l.Error(
			"发送不一致消息失败",
			logger.Error(err),
			logger.String("type", typ),
			logger.Int64("id", id),
		)
	}
}

func (v *Validator[T]) Full() *Validator[T] {
	v.fromBase = v.fullFromBase
	return v
}

func (v *Validator[T]) Incr() *Validator[T] {
	v.fromBase = v.incrFromBase
	return v
}

func (v *Validator[T]) UpdateTime(t int64) *Validator[T] {
	v.updateTime = t
	return v
}

func (v *Validator[T]) SleepInterval(interval time.Duration) *Validator[T] {
	v.sleepInterval = interval
	return v
}

// 全量
func (v *Validator[T]) fullFromBase(ctx context.Context, offset int) ([]T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src []T
	err := v.base.WithContext(dbCtx).Order("id").
		Offset(offset).Limit(v.batchSize).Find(&src).Error
	return src, err
}

// 增量
func (v *Validator[T]) incrFromBase(ctx context.Context, offset int) ([]T, error) {
	dbCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var src []T
	err := v.base.WithContext(dbCtx).
		Where("update_time > ?", v.updateTime).
		Order("update_time").
		Offset(offset).Limit(v.batchSize).
		Find(&src).Error
	return src, err
}
