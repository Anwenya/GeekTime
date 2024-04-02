package oom

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	_ "net/http/pprof"
	"os"
	"testing"
)

func TestServer(t *testing.T) {
	engine := gin.Default()
	engine.GET("/readAll", testReadAll)
	engine.GET("/ioCopy", testIOCopy)
	err := engine.Run(":8090")
	require.NoError(t, err)
}

func testReadAll(ctx *gin.Context) {
	file, err := os.Open("C:\\Users\\Administrator\\Desktop\\takeout-20240327T091209Z-001.zip")
	if err != nil {
		panic(any(err))
	}
	defer file.Close()

	// 这个操作会一次性将文件读取到内存中
	bytes, err := io.ReadAll(file)
	log.Printf("readAll:%d", len(bytes))
	if err != nil {
		panic(any(err))
	}

	_, err = ctx.Writer.Write(bytes)
	if err != nil {
		panic(any(err))
	}
}

func testIOCopy(ctx *gin.Context) {
	file, err := os.Open("C:\\Users\\Administrator\\Desktop\\takeout-20240327T091209Z-001.zip")
	if err != nil {
		panic(any(err))
	}
	defer file.Close()

	// 流拷贝 默认buf是32kb
	count, err := io.Copy(ctx.Writer, file)
	log.Printf("ioCopy:%d", count)
	if err != nil {
		panic(any(err))
	}
}
