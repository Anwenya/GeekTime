package middleware

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type LogMiddlewareBuilder struct {
	logFunc       func(ctx context.Context, al AccessLog)
	allowReqBody  bool
	allowRespBody bool
}

func NewLogMiddlewareBuilder(logFunc func(ctx context.Context, al AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFunc: logFunc,
	}
}

func (lmb *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	lmb.allowReqBody = true
	return lmb
}

func (lmb *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	lmb.allowRespBody = true
	return lmb
}

func (lmb *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		// 做一下长度限制
		if len(path) > 1024 {
			path = path[:1024]
		}
		method := ctx.Request.Method
		al := AccessLog{
			Path:   path,
			Method: method,
		}

		if lmb.allowReqBody {
			// ctx.Request.Body是stream对象
			// stream对象只能读一次 事后需要再创建一个放回去
			body, _ := ctx.GetRawData()
			if len(body) > 2048 {
				al.ReqBody = string(body[:2048])
			} else {
				al.ReqBody = string(body)
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		// 如果记录响应体就替换一下装饰器
		if lmb.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				al:             &al,
			}
		}
		// 记录耗时
		start := time.Now()
		defer func() {
			al.Duration = time.Since(start)
			lmb.logFunc(ctx, al)
		}()

		ctx.Next()
	}
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	Status   int           `json:"status"`
	RespBody string        `json:"resp_body"`
	Duration time.Duration `json:"duration"`
}

// 装饰原有的ResponseWriter
// 来记录需要的内容到日志中
type responseWriter struct {
	gin.ResponseWriter
	al *AccessLog
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.al.RespBody = string(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.al.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
