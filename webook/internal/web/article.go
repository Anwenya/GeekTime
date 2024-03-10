package web

import (
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx"
	"github.com/Anwenya/GeekTime/webook/pkg/ginx/decorator"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	articleService     service.ArticleService
	interactiveService service.InteractiveService
	l                  logger.LoggerV1
	biz                string
}

func NewArticleHandler(
	l logger.LoggerV1,
	articleService service.ArticleService,
	interactiveService service.InteractiveService,
) *ArticleHandler {
	return &ArticleHandler{
		l:                  l,
		articleService:     articleService,
		interactiveService: interactiveService,
		biz:                "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("/edit", decorator.WrapBodyAndClaims[ArticleEditReq, token.UserClaims](h.Edit))
	group.POST("/publish", decorator.WrapBodyAndClaims[PublishReq, token.UserClaims](h.Publish))
	group.POST("/withdraw", decorator.WrapBodyAndClaims[ArticleWithdrawReq, token.UserClaims](h.Withdraw))
	group.GET("/detail/:id", h.Detail)
	group.POST("/list", h.List)

	pubGroup := group.Group("/pub")
	pubGroup.GET("/:id", h.PubDetail)

	pubGroup.POST("/like", decorator.WrapBodyAndClaims[ArticleLikeReq, token.UserClaims](h.Like))
	pubGroup.POST("/collect", decorator.WrapBodyAndClaims[ArticleCollectReq, token.UserClaims](h.LikeCollect))

}

func (h *ArticleHandler) Edit(
	ctx *gin.Context,
	req ArticleEditReq,
	uc token.UserClaims,
) (ginx.Result, error) {
	id, err := h.articleService.Save(
		ctx,
		domain.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: domain.Author{
				Id: uc.Uid,
			},
		},
	)
	if err != nil {
		return ginx.Result{Msg: "系统错误"}, err
	}
	return ginx.Result{Data: id}, nil
}

func (h *ArticleHandler) Publish(
	ctx *gin.Context,
	req PublishReq,
	uc token.UserClaims,
) (ginx.Result, error) {

	id, err := h.articleService.Publish(
		ctx,
		domain.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: domain.Author{
				Id: uc.Uid,
			},
		},
	)
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}

	return ginx.Result{Data: id}, nil
}

func (h *ArticleHandler) Withdraw(
	ctx *gin.Context,
	req ArticleWithdrawReq,
	uc token.UserClaims,
) (ginx.Result, error) {
	err := h.articleService.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}

	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Msg:  "id 参数错误",
				Code: 4,
			},
		)
		h.l.Warn(
			"查询文章失败 id 格式异常",
			logger.String("id", idstr),
			logger.Error(err),
		)
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	art, err := h.articleService.GetById(ctx, id)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Msg:  "系统错误",
				Code: 5,
			},
		)
		h.l.Error(
			"查询文章失败",
			logger.Int64("id", id),
			logger.Error(err),
		)
		return
	}

	if art.Author.Id != uc.Uid {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Msg:  "系统错误",
				Code: 5,
			},
		)
		h.l.Error(
			"非法查询文章",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid),
		)
		return
	}

	vo := ArticleVo{
		Id:         art.Id,
		Title:      art.Title,
		Content:    art.Content,
		AuthorId:   art.Author.Id,
		Status:     art.Status.ToUint8(),
		CreateTime: art.CreateTime.Format(time.DateTime),
		UpdateTime: art.UpdateTime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, ginx.Result{Data: vo})

}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.BindJSON(&page); err != nil {
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	arts, err := h.articleService.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		)
		h.l.Error(
			"查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		ginx.Result{
			Data: slice.Map[domain.Article, ArticleVo](
				arts,
				func(idx int, src domain.Article) ArticleVo {
					return ArticleVo{
						Id:       src.Id,
						Title:    src.Title,
						Abstract: src.Abstract(),
						AuthorId: src.Author.Id,
						// 列表接口可能不需要该字段 以实际业务需求灵活变动
						Status:     src.Status.ToUint8(),
						CreateTime: src.CreateTime.Format(time.DateTime),
						UpdateTime: src.UpdateTime.Format(time.DateTime),
					}
				},
			),
		},
	)
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Msg:  "id 参数错误",
				Code: 4,
			},
		)
		h.l.Warn(
			"查询文章失败 id 格式异常",
			logger.String("id", idStr),
			logger.Error(err),
		)
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)

	var (
		eg          errgroup.Group
		art         domain.Article
		interactive domain.Interactive
	)
	eg.Go(func() error {
		var er error
		art, er = h.articleService.GetPubById(ctx, id, uc.Uid)
		return er
	})

	eg.Go(func() error {
		var er error
		interactive, er = h.interactiveService.Get(ctx, h.biz, id, uc.Uid)
		return er
	})

	if err != nil {
		ctx.JSON(
			http.StatusOK,
			ginx.Result{
				Msg:  "系统错误",
				Code: 5,
			},
		)
		h.l.Error(
			"查询文章失败 系统错误",
			logger.Int64("aid", art.Id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err),
		)
		return
	}
	// 更改为异步消息
	//go func() {
	//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//
	//	err := h.interactiveService.IncrReadCnt(ctx, h.biz, art.Id)
	//	if err != nil {
	//		h.l.Error(
	//			"更新阅读数失败",
	//			logger.Int64("aid", art.Id),
	//			logger.Error(err),
	//		)
	//	}
	//}()

	ctx.JSON(
		http.StatusOK,
		ginx.Result{
			Data: ArticleVo{
				Id:         art.Id,
				Title:      art.Title,
				Content:    art.Content,
				AuthorId:   art.Author.Id,
				AuthorName: art.Author.Name,

				ReadCnt:    interactive.ReadCnt,
				CollectCnt: interactive.CollectCnt,
				LikeCnt:    interactive.LikeCnt,
				Liked:      interactive.Liked,
				Collected:  interactive.Collected,

				Status:     art.Status.ToUint8(),
				CreateTime: art.CreateTime.Format(time.DateTime),
				UpdateTime: art.UpdateTime.Format(time.DateTime),
			},
		},
	)
}

func (h *ArticleHandler) Like(
	ctx *gin.Context,
	req ArticleLikeReq,
	uc token.UserClaims,
) (ginx.Result, error) {
	var err error
	if req.Like {
		err = h.interactiveService.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.interactiveService.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}

	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Result{Msg: "OK"}, nil
}

func (h *ArticleHandler) LikeCollect(
	ctx *gin.Context,
	req ArticleCollectReq,
	uc token.UserClaims,
) (ginx.Result, error) {
	err := h.interactiveService.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{Msg: "OK"}, nil
}
