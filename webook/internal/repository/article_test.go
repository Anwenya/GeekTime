package repository

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/repository/dao"
	mockdao "github.com/Anwenya/GeekTime/webook/internal/repository/dao/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCachedArticleRepository_SyncV1(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO)
		art     domain.Article
		wantId  int64
		wantErr error
	}{
		{
			name: "新建同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				adao := mockdao.NewMockArticleAuthorDAO(ctrl)
				adao.EXPECT().
					Create(
						gomock.Any(),
						dao.Article{
							Title:    "标题",
							Content:  "内容",
							AuthorId: 123,
						},
					).Return(int64(1), nil)
				rdao := mockdao.NewMockArticleReaderDAO(ctrl)
				rdao.EXPECT().
					UpsertV2(
						gomock.Any(),
						dao.Article{
							Id:       1,
							Title:    "标题",
							Content:  "内容",
							AuthorId: 123,
						},
					).Return(nil)
				return adao, rdao
			},
			art: domain.Article{
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 1,
		},
		{
			name: "修改同步成功",
			mock: func(ctrl *gomock.Controller) (dao.ArticleAuthorDAO, dao.ArticleReaderDAO) {
				adao := mockdao.NewMockArticleAuthorDAO(ctrl)
				adao.EXPECT().
					Update(
						gomock.Any(),
						dao.Article{
							Id:       11,
							Title:    "我的标题",
							Content:  "我的内容",
							AuthorId: 123,
						},
					).Return(nil)
				rdao := mockdao.NewMockArticleReaderDAO(ctrl)
				rdao.EXPECT().
					UpsertV2(
						gomock.Any(),
						dao.Article{
							Id:       11,
							Title:    "我的标题",
							Content:  "我的内容",
							AuthorId: 123,
						},
					).Return(nil)
				return adao, rdao
			},
			art: domain.Article{
				Id:      11,
				Title:   "我的标题",
				Content: "我的内容",
				Author: domain.Author{
					Id: 123,
				},
			},
			wantId: 11,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			authorDAO, readerDAO := tc.mock(ctrl)
			repo := NewCachedArticleRepositoryV2(readerDAO, authorDAO)
			id, err := repo.SyncV1(context.Background(), tc.art)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantId, id)
		})
	}
}
