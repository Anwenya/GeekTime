package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

var errUnknownPattern = errors.New("未知的双写模式")

const (
	PatternSrcOnly  = "src_only"
	PatternSrcFirst = "src_first"
	PatternDstFirst = "dst_first"
	PatternDstOnly  = "dst_only"
)

type DoubleWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[string]
	l       logger.LoggerV1
}

func NewDoubleWritePool(src *gorm.DB, dst *gorm.DB, l logger.LoggerV1) *DoubleWritePool {
	return &DoubleWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomicx.NewValueOf[string](PatternSrcOnly),

		l: l,
	}
}

// UpdatePattern 更新双写模式
func (d *DoubleWritePool) UpdatePattern(pattern string) error {
	switch pattern {
	case PatternSrcOnly, PatternSrcFirst,
		PatternDstOnly, PatternDstFirst:
		d.pattern.Store(pattern)
		return nil
	default:
		return errUnknownPattern
	}
}

// BeginTx 开启事务
func (d *DoubleWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTX{
			src:     src,
			pattern: pattern,
			l:       d.l,
		}, err
	case PatternSrcFirst:
		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}

		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写目标表开启事务失败", logger.Error(err))
		}

		return &DoubleWriteTX{
			src:     src,
			dst:     dst,
			pattern: pattern,
			l:       d.l,
		}, nil
	case PatternDstOnly:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		return &DoubleWriteTX{
			dst:     dst,
			pattern: pattern,
			l:       d.l,
		}, err
	case PatternDstFirst:
		dst, err := d.dst.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}

		src, err := d.src.(gorm.TxBeginner).BeginTx(ctx, opts)
		if err != nil {
			d.l.Error("双写源表开启事务失败", logger.Error(err))
		}
		return &DoubleWriteTX{
			src:     src,
			dst:     dst,
			pattern: pattern,
			l:       d.l,
		}, nil
	}
}

func (d *DoubleWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic(any("双写模式写不支持"))
}

func (d *DoubleWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args)

		if err != nil || d.dst == nil {
			return res, err
		}

		_, err = d.dst.ExecContext(ctx, query, args)
		if err != nil {
			d.l.Error(
				"双写写入 dst 失败", logger.Error(err),
				logger.String("sql", query),
			)
		}

		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args)
		if err != nil || d.src == nil {
			return res, err
		}

		_, err = d.src.ExecContext(ctx, query, args)
		if err != nil {
			d.l.Error(
				"双写写入 src 失败",
				logger.Error(err),
				logger.String("sql", query),
			)
		}
		return res, nil
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args)
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args)
	default:
		// sql.Row 的字段都私有的
		// 正常手段这了返回不了异常信息
		// 极端的做法可以直接panic
		panic(any(errUnknownPattern))
	}
}

// DoubleWriteTX 处理事务相关
type DoubleWriteTX struct {
	src     *sql.Tx
	dst     *sql.Tx
	pattern string
	l       logger.LoggerV1
}

func (d *DoubleWriteTX) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		// 如果源表都提交失败了 就不要再管目标表了
		if err != nil {
			return err
		}

		if d.dst != nil {
			return nil
		}

		err = d.dst.Commit()
		if err != nil {
			d.l.Error("目标表提交事务失败")
		}

		return nil
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}

		if d.src == nil {
			return nil
		}

		err = d.src.Commit()

		if err != nil {
			d.l.Error("源表提交事务失败")
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteTX) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}

		if d.dst == nil {
			return nil
		}

		err = d.dst.Rollback()
		if err != nil {
			d.l.Error("目标表回滚失败")
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}

		if d.src == nil {
			return nil
		}

		err = d.src.Rollback()
		if err != nil {
			d.l.Error("源表回滚失败")
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DoubleWriteTX) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	panic(any("双写模式写不支持"))
}

func (d *DoubleWriteTX) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args)

		if err != nil || d.dst == nil {
			return res, err
		}

		_, err = d.dst.ExecContext(ctx, query, args)
		if err != nil {
			d.l.Error(
				"双写写入 dst 失败", logger.Error(err),
				logger.String("sql", query),
			)
		}

		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args)
		if err != nil || d.src == nil {
			return res, err
		}

		_, err = d.src.ExecContext(ctx, query, args)
		if err != nil {
			d.l.Error(
				"双写写入 src 失败",
				logger.Error(err),
				logger.String("sql", query),
			)
		}
		return res, nil
	default:
		return nil, errUnknownPattern
	}
}

func (d *DoubleWriteTX) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args)
	default:
		return nil, errUnknownPattern
	}

}

func (d *DoubleWriteTX) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args)
	default:
		// sql.Row 的字段都私有的
		// 正常手段这了返回不了异常信息
		// 极端的做法可以直接panic
		panic(any(errUnknownPattern))
	}
}
