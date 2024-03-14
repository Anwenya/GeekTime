package cronjob

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	ticker := time.NewTicker(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Log("循环结束")
			goto end
		case now := <-ticker.C:
			t.Log(now.UnixMilli())
		}
	}

end:
	t.Log("结束")
}
