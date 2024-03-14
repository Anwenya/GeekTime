package cronjob

import (
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCronExpr(t *testing.T) {
	expr := cron.New(cron.WithSeconds())

	id, err := expr.AddFunc("@every 1s", func() {
		t.Log("执行了")
	})

	assert.NoError(t, err)
	t.Log("任务", id)
	expr.Start()
	time.Sleep(time.Second * 10)
	// 停止调度新任务 但正在执行的任务会执行完成
	ctx := expr.Stop()
	t.Log("停止信号")
	<-ctx.Done()
	t.Log("停止")
}
