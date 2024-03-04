package integration

import (
	"bytes"
	"encoding/json"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/integration/startup"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 测试套件
type ArticleGORMHandlerSuite struct {
	suite.Suite
	db     *gorm.DB
	server *gin.Engine
}

// which will run before the tests in the suite are run
func (s *ArticleGORMHandlerSuite) SetupSuite() {
	l := startup.InitLogger()
	// 拿到db才能对结果进行校验等操作
	s.db = startup.InitDB(l)
	gormDAO := dao.NewGORMArticleDAO(s.db)
	h := startup.InitArticleHandler(gormDAO)
	server := gin.Default()
	// 用于登录校验
	server.Use(
		func(ctx *gin.Context) {
			ctx.Set(
				"user",
				token.UserClaims{
					Uid: 111,
				},
			)
		},
	)
	h.RegisterRoutes(server)
	s.server = server
}

// which will run after each test in the suite.
func (s *ArticleGORMHandlerSuite) TearDownTest() {
	// 自增id也会被重置
	err := s.db.Exec("truncate table `articles`").Error
	assert.NoError(s.T(), err)
	err = s.db.Exec("truncate table `published_articles`").Error
	assert.NoError(s.T(), err)
}

// 测试用例
func (s *ArticleGORMHandlerSuite) TestArticlePublish() {
	t := s.T()

	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		req Article

		wantCode   int
		wantResult Result[int64]
	}{
		{
			name: "新建帖子并发表",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.Article
				s.db.Where("author_id = ?", 111).First(&art)
				assert.Equal(t, "新建帖子并发表的标题", art.Title)
				assert.Equal(t, "新建帖子并发表的内容", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), art.Status)
				assert.Equal(t, int64(111), art.AuthorId)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)

				var pa dao.PublishedArticle
				s.db.Where("author_id = ?", 111).First(&pa)
				assert.Equal(t, "新建帖子并发表的标题", pa.Title)
				assert.Equal(t, "新建帖子并发表的内容", pa.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), pa.Status)
				assert.Equal(t, int64(111), pa.AuthorId)
				assert.True(t, pa.CreateTime > 0)
				assert.True(t, pa.UpdateTime > 0)
			},
			req: Article{
				Title:   "新建帖子并发表的标题",
				Content: "新建帖子并发表的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "更新帖子并发表",
			before: func(t *testing.T) {
				s.db.Create(
					&dao.Article{
						Id:         2,
						AuthorId:   111,
						Title:      "更新帖子并发表的标题",
						Content:    "更新帖子并发表的内容",
						Status:     uint8(domain.ArticleStatusUnpublished),
						CreateTime: 123,
						UpdateTime: 123,
					},
				)
			},
			after: func(t *testing.T) {
				var art dao.Article
				s.db.Where("id = ?", 2).First(&art)
				assert.Equal(t, "新的更新帖子并发表的标题", art.Title)
				assert.Equal(t, "新的更新帖子并发表的内容", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), art.Status)
				assert.Equal(t, int64(111), art.AuthorId)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)

				var pa dao.PublishedArticle
				s.db.Where("id = ?", 2).First(&pa)
				assert.Equal(t, "新的更新帖子并发表的标题", pa.Title)
				assert.Equal(t, "新的更新帖子并发表的内容", pa.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), pa.Status)
				assert.Equal(t, int64(111), pa.AuthorId)
				assert.True(t, pa.CreateTime > 0)
				assert.True(t, pa.UpdateTime > 0)
			},
			req: Article{
				Id:      2,
				Title:   "新的更新帖子并发表的标题",
				Content: "新的更新帖子并发表的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "更新帖子并重新发表",
			before: func(t *testing.T) {
				s.db.Create(
					&dao.Article{
						Id:         3,
						AuthorId:   111,
						Title:      "更新帖子并重新发表的标题",
						Content:    "更新帖子并重新发表的内容",
						Status:     uint8(domain.ArticleStatusPublished),
						CreateTime: 123,
						UpdateTime: 123,
					},
				)

				s.db.Create(
					&dao.PublishedArticle{
						Id:         3,
						AuthorId:   111,
						Title:      "更新帖子并重新发表的标题",
						Content:    "更新帖子并重新发表的内容",
						Status:     uint8(domain.ArticleStatusPublished),
						CreateTime: 123,
						UpdateTime: 123,
					},
				)
			},
			after: func(t *testing.T) {
				var art dao.Article
				s.db.Where("id = ?", 3).First(&art)
				assert.Equal(t, "新的更新帖子并重新发表的标题", art.Title)
				assert.Equal(t, "新的更新帖子并重新发表的内容", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), art.Status)
				assert.Equal(t, int64(111), art.AuthorId)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)

				var pa dao.PublishedArticle
				s.db.Where("id = ?", 3).First(&pa)
				assert.Equal(t, "新的更新帖子并重新发表的标题", pa.Title)
				assert.Equal(t, "新的更新帖子并重新发表的内容", pa.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), pa.Status)
				assert.Equal(t, int64(111), pa.AuthorId)
				assert.True(t, pa.CreateTime > 0)
				assert.True(t, pa.UpdateTime > 0)
			},
			req: Article{
				Id:      3,
				Title:   "新的更新帖子并重新发表的标题",
				Content: "新的更新帖子并重新发表的内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Data: 3,
			},
		},
		{
			name: "更新别人的帖子并发表-失败",
			before: func(t *testing.T) {
				s.db.Create(
					&dao.Article{
						Id:         4,
						AuthorId:   222,
						Title:      "别人的帖子",
						Content:    "别人的帖子",
						Status:     uint8(domain.ArticleStatusPublished),
						CreateTime: 123,
						UpdateTime: 123,
					},
				)

				s.db.Create(
					&dao.PublishedArticle{
						Id:         4,
						AuthorId:   222,
						Title:      "别人的帖子",
						Content:    "别人的帖子",
						Status:     uint8(domain.ArticleStatusPublished),
						CreateTime: 123,
						UpdateTime: 123,
					},
				)
			},
			after: func(t *testing.T) {
				var art dao.Article
				s.db.Where("id = ?", 4).First(&art)
				assert.Equal(t, "别人的帖子", art.Title)
				assert.Equal(t, "别人的帖子", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), art.Status)
				assert.Equal(t, int64(222), art.AuthorId)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)

				var pa dao.PublishedArticle
				s.db.Where("id = ?", 4).First(&pa)
				assert.Equal(t, "别人的帖子", pa.Title)
				assert.Equal(t, "别人的帖子", pa.Content)
				assert.Equal(t, uint8(domain.ArticleStatusPublished), pa.Status)
				assert.Equal(t, int64(222), pa.AuthorId)
				assert.True(t, pa.CreateTime > 0)
				assert.True(t, pa.UpdateTime > 0)
			},
			req: Article{
				Id:      4,
				Title:   "更新内容",
				Content: "更新内容",
			},
			wantCode: 200,
			wantResult: Result[int64]{
				Code: 5,
				Msg:  "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			data, err := json.Marshal(tc.req)
			assert.NoError(t, err)
			req, err := http.NewRequest(
				http.MethodPost,
				"/articles/publish",
				bytes.NewReader(data),
			)
			assert.NoError(t, err)
			req.Header.Set(
				"Content-Type",
				"application/json",
			)

			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)

			if code != http.StatusOK {
				return
			}

			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantResult, result)
		})
	}
}

// 测试用例
func (s *ArticleGORMHandlerSuite) TestEdit() {
	t := s.T()

	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		art      Article
		wantCode int
		wantRes  Result[int64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				var art dao.Article

				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				assert.Equal(t, "标题", art.Title)
				assert.Equal(t, "内容", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusUnpublished), art.Status)
				assert.Equal(t, int64(111), art.AuthorId)
			},
			art: Article{
				Title:   "标题",
				Content: "内容",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 1,
			},
		},
		{
			name: "修改帖子",
			before: func(t *testing.T) {
				s.db.Create(&dao.Article{
					Id:         2,
					AuthorId:   111,
					Title:      "原帖子",
					Content:    "原帖子",
					Status:     uint8(domain.ArticleStatusUnpublished),
					CreateTime: 123,
					UpdateTime: 123,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article

				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				assert.Equal(t, "新帖子", art.Title)
				assert.Equal(t, "新帖子", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusUnpublished), art.Status)
				assert.Equal(t, int64(111), art.AuthorId)
			},
			art: Article{
				Id:      2,
				Title:   "新帖子",
				Content: "新帖子",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Data: 2,
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				s.db.Create(&dao.Article{
					Id:         3,
					AuthorId:   222,
					Title:      "别人的帖子",
					Content:    "别人的帖子",
					Status:     uint8(domain.ArticleStatusUnpublished),
					CreateTime: 123,
					UpdateTime: 123,
				})
			},
			after: func(t *testing.T) {
				var art dao.Article

				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.CreateTime > 0)
				assert.True(t, art.UpdateTime > 0)
				assert.Equal(t, "别人的帖子", art.Title)
				assert.Equal(t, "别人的帖子", art.Content)
				assert.Equal(t, uint8(domain.ArticleStatusUnpublished), art.Status)
				assert.Equal(t, int64(222), art.AuthorId)
			},
			art: Article{
				Id:      3,
				Title:   "新帖子",
				Content: "新帖子",
			},
			wantCode: http.StatusOK,
			wantRes: Result[int64]{
				Msg: "系统错误",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)

			data, err := json.Marshal(&tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(
				http.MethodPost,
				"/articles/edit",
				bytes.NewReader(data),
			)
			assert.NoError(t, err)

			req.Header.Set(
				"Content-Type",
				"application/json",
			)

			recorder := httptest.NewRecorder()
			s.server.ServeHTTP(recorder, req)
			code := recorder.Code
			assert.Equal(t, tc.wantCode, code)

			if code != http.StatusOK {
				return
			}
			var result Result[int64]
			err = json.Unmarshal(recorder.Body.Bytes(), &result)
			assert.NoError(t, err)
			assert.Equal(t, tc.wantRes, result)
		})
	}
}

// 入口
func TestArticleHandler(t *testing.T) {
	suite.Run(t, &ArticleGORMHandlerSuite{})
}

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

type Article struct {
	Id      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}
