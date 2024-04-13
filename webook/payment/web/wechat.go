package web

import (
	"github.com/Anwenya/GeekTime/webook/payment/service/wechat"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"net/http"
)

type WechatHandler struct {
	handler   *notify.Handler
	nativeSvc *wechat.NativePaymentService
	l         logger.LoggerV1
}

func NewWechatHandler(
	handler *notify.Handler,
	nativeSvc *wechat.NativePaymentService,
	l logger.LoggerV1,
) *WechatHandler {
	return &WechatHandler{
		handler:   handler,
		nativeSvc: nativeSvc,
		l:         l,
	}
}

func (h *WechatHandler) RegisterRoutes(server *gin.Engine) {
	server.Any("/pay/callback", h.HandleNative)
}

func (h *WechatHandler) HandleNative(ctx *gin.Context) {
	//
	transaction := new(payments.Transaction)
	// 从 HTTP 请求(http.Request) 中解析 微信支付通知
	_, err := h.handler.ParseNotifyRequest(ctx, ctx.Request, transaction)
	if err != nil {
		// 第三方支付的稳定性应该是很高的
		// 如果有异常大概率是非法攻击
		ctx.String(http.StatusBadRequest, "参数解析失败")
		h.l.Error("解析微信支付回调失败", logger.Error(err))
		return
	}
	// 支付消息可以发送到kafka
	// 其他关心该动作的服务可以使用
	err = h.nativeSvc.HandleCallback(ctx, transaction)
	if err != nil {
		// 处理回调失败 主动发起对账
		_ = h.nativeSvc.SyncWechatInfo(ctx, *transaction.OutTradeNo)

		ctx.String(http.StatusInternalServerError, "系统异常")
		h.l.Error(
			"处理微信支付回调失败",
			logger.Error(err),
			logger.String("biz_trade_no", *transaction.OutTradeNo),
		)
		return
	}
	ctx.String(http.StatusOK, "OK")
}
