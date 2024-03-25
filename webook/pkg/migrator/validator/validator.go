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

	l logger.LoggerV1
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
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var bs []T
		// 一次查询出一批 然后逐个校验
		err := v.base.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&bs).Error

		// 无结果 结束
		if err == gorm.ErrRecordNotFound || len(bs) == 0 {
			return nil
		}

		// 失败跳过
		if err != nil {
			v.l.Error(
				"base => target 查询 base",
				logger.Error(err),
			)
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
		err = v.target.WithContext(ctx).Select("id").
			Where("id IN ?", ids).Find(&dstTs).Error

		// 目标表没有数据
		if err == gorm.ErrRecordNotFound || len(dstTs) == 0 {
			// 生成对应的修复任务
			v.notifyTargetMissing(bs)
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
			return nil
		}
	}
}

// 以 target 为准
func (v *Validator[T]) validateTarget2Base(ctx context.Context) error {
	offset := -v.batchSize
	for {
		offset += v.batchSize
		var ts []T
		// 一次查询出一批 然后逐个校验
		err := v.target.WithContext(ctx).Select("id").
			Order("id").Offset(offset).Limit(v.batchSize).
			Find(&ts).Error

		// 无结果 结束
		if err == gorm.ErrRecordNotFound || len(ts) == 0 {
			return nil
		}

		// 失败跳过
		if err != nil {
			v.l.Error(
				"target => base 查询 target失败",
				logger.Error(err),
			)
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
			return nil
		}
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
	v.l.Error(
		"发送不一致消息失败",
		logger.Error(err),
		logger.String("type", typ),
		logger.Int64("id", id),
	)
}
