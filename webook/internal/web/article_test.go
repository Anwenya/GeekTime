package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	mockservice "github.com/Anwenya/GeekTime/webook/internal/service/mocks"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string
		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody  Article
		wantCode int
		wantRes  Result
	}{
		{
			name: "新建并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mockservice.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(),
					domain.Article{
						Title:   "我的标题",
						Content: "我的内容",
						Author: domain.Author{
							Id: 123,
						},
					}).Return(int64(1), nil)
				return svc
			},
			reqBody: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: 200,
			wantRes: Result{
				// 原本是 int64的 但是因为 Data 是any
				// 所以在反序列化的时候默认用的 float64
				Data: float64(1),
			},
		},
		{
			name: "修改并且发表成功",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mockservice.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(),
					domain.Article{
						Id:      1,
						Title:   "新的标题",
						Content: "新的内容",
						Author: domain.Author{
							Id: 123,
						},
					}).Return(int64(1), nil)
				return svc
			},
			reqBody: Article{
				Id:      1,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: Result{
				Data: float64(1),
			},
		},
		{
			name: "输入有误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mockservice.NewMockArticleService(ctrl)
				return svc
			},
			reqBody:  Article{},
			wantCode: 400,
		},
		{
			name: "publish错误",
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := mockservice.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(),
					domain.Article{
						Id:      1,
						Title:   "新的标题",
						Content: "新的内容",
						Author: domain.Author{
							Id: 123,
						},
					}).Return(int64(0), errors.New("mock error"))
				return svc
			},
			reqBody: Article{
				Id:      1,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: 200,
			wantRes: Result{
				Msg:  "系统错误",
				Code: 5,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := tc.mock(ctrl)
			hdl := NewArticleHandler(logger.NewNopLogger(), svc)

			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set(
					"user",
					token.UserClaims{
						Uid: 123,
					})
			})
			hdl.RegisterRoutes(server)

			var data []byte
			var err error
			// 伪造异常的数据
			if tc.name == "输入有误" {
				data = []byte("123")
			} else {
				data, err = json.Marshal(tc.reqBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest(
				http.MethodPost,
				"/articles/publish",
				bytes.NewReader(data),
			)
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if recorder.Code != http.StatusOK {
				return
			}
			var res Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
