package fixer

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator"
	"github.com/Anwenya/GeekTime/webook/pkg/migrator/events"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OverrideFixer[T migrator.Entity] struct {
	base   *gorm.DB
	target *gorm.DB

	// 修复的字段/列
	columns []string
}

func NewOverrideFixer[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
) (*OverrideFixer[T], error) {
	rows, err := base.Model(new(T)).Order("id").Rows()
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	return &OverrideFixer[T]{base: base, target: target, columns: columns}, err

}

func (o *OverrideFixer[T]) Fix(ctx context.Context, id int64) error {
	// 不校验是否相等 直接覆盖 最为简单粗暴的写法
	var t T
	err := o.base.WithContext(ctx).Where("id = ?", id).First(&t).Error
	switch err {
	case gorm.ErrRecordNotFound:
		// 源表没有该记录 目标表也删除
		return o.target.WithContext(ctx).Model(&t).Delete("id = ?", id).Error
	case nil:
		// 源表中有该记录 执行插入/更新语句 有则更新/无则插入
		return o.target.WithContext(ctx).
			Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(o.columns),
			},
			).Create(&t).Error
	default:
		// 其他异常 直接返回
		return err
	}
}

// NewOverrideFixerV1 同上 但列由外部传入
func NewOverrideFixerV1[T migrator.Entity](
	base *gorm.DB,
	target *gorm.DB,
	columns []string,
) *OverrideFixer[T] {
	return &OverrideFixer[T]{
		base:    base,
		target:  target,
		columns: columns}
}

func (o *OverrideFixer[T]) FixV1(evt events.InconsistentEvent) error {
	switch evt.Type {
	case events.InconsistentEventTypeNEQ,
		events.InconsistentEventTypeTargetMissing:
		var t T
		err := o.base.Where("id = ?", evt.ID).First(&t).Error
		switch err {
		case gorm.ErrRecordNotFound:
			// 源表没有该记录 目标表也删除
			return o.target.Model(&t).Delete("id = ?", evt.ID).Error
		case nil:
			// 源表中有该记录 执行插入/更新语句 有则更新/无则插入
			return o.target.Clauses(clause.OnConflict{
				DoUpdates: clause.AssignmentColumns(o.columns),
			},
			).Create(&t).Error
		default:
			// 其他异常 直接返回
			return err
		}
	case events.InconsistentEventTypeBaseMissing:
		return o.target.Model(new(T)).Delete("id = ?", evt.ID).Error

	}
	return nil
}
