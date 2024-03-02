package web

import (
	"github.com/Anwenya/GeekTime/webook/internal/domain"
	"github.com/Anwenya/GeekTime/webook/internal/service"
	"github.com/Anwenya/GeekTime/webook/internal/web/token"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type ArticleHandler struct {
	svc service.ArticleService
	l   logger.LoggerV1
}

func NewArticleHandler(
	l logger.LoggerV1,
	svc service.ArticleService,
) *ArticleHandler {
	return &ArticleHandler{
		l:   l,
		svc: svc,
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	group := server.Group("/articles")
	group.POST("/edit", h.Edit)
	group.POST("/publish", h.Publish)
	group.POST("/withdraw", h.Withdraw)
	group.GET("/detail/:id", h.Detail)
	group.POST("/list", h.List)

	pubGroup := group.Group("/pub")
	pubGroup.GET("/:id", h.PubDetail)

}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.l.Error("保存文章数据失败",
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	id, err := h.svc.Publish(
		ctx,
		domain.Article{
			Id:      req.Id,
			Title:   req.Title,
			Content: req.Content,
			Author: domain.Author{
				Id: uc.Uid,
			},
		})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("发表文章失败",
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(token.UserClaims)
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		ctx.JSON(
			http.StatusOK,
			Result{
				Msg:  "系统错误",
				Code: 5,
			},
		)
		h.l.Error("撤回文章失败",
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id),
			logger.Error(err))
		return
	}
	ctx.JSON(
		http.StatusOK,
		Result{
			Msg: "OK",
		},
	)
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn(
			"查询文章失败 id 格式异常",
			logger.String("id", idstr),
			logger.Error(err),
		)
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error(
			"查询文章失败",
			logger.Int64("id", id),
			logger.Error(err),
		)
		return
	}

	if art.Author.Id != uc.Uid {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
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
	ctx.JSON(http.StatusOK, Result{Data: vo})

}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.BindJSON(&page); err != nil {
		return
	}

	uc := ctx.MustGet("user").(token.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid),
		)
		return
	}

	ctx.JSON(
		http.StatusOK,
		Result{
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
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)

	art, err := h.svc.GetPubById(ctx, id)

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn(
			"查询文章失败 id 格式异常",
			logger.String("id", idstr),
			logger.Error(err),
		)
		return
	}

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error(
			"查询文章失败 系统错误",
			logger.Error(err),
		)
		return
	}

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Status:     art.Status.ToUint8(),
			CreateTime: art.CreateTime.Format(time.DateTime),
			UpdateTime: art.UpdateTime.Format(time.DateTime),
		},
	})
}
