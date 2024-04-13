package wechat

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/payment/domain"
	"github.com/Anwenya/GeekTime/webook/payment/events"
	"github.com/Anwenya/GeekTime/webook/payment/repository"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
)

type NativePaymentService struct {
	appId string
	mchId string
	// 支付回调 url
	notifyUrl string
	// 同步在本地做的支付记录
	repo repository.PaymentRepository

	svc      *native.NativeApiService
	producer events.Producer

	// 回调类型状态映射
	nativeCBTypeToStatus map[string]domain.PaymentStatus

	l logger.LoggerV1
}

func NewNativePaymentService(
	appId string, mchId string,
	repo repository.PaymentRepository,
	svc *native.NativeApiService,
	producer events.Producer,
	l logger.LoggerV1,
) *NativePaymentService {
	return &NativePaymentService{
		appId:     appId,
		mchId:     mchId,
		repo:      repo,
		svc:       svc,
		notifyUrl: "http://localhost:8070/pay/callback",
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
		},
		producer: producer,
		l:        l,
	}
}

func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	// 初始状态
	pmt.Status = domain.PaymentStatusInit
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}

	// 向微信发起支付请求
	resp, _, err := n.svc.Prepay(
		ctx,
		native.PrepayRequest{
			Appid:       core.String(n.appId),
			Mchid:       core.String(n.mchId),
			Description: core.String(pmt.Description),
			OutTradeNo:  core.String(pmt.BizTradeNO),
			// 支付有效期
			TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
			Amount: &native.Amount{
				Total:    core.Int64(pmt.Amt.Total),
				Currency: core.String(pmt.Amt.Currency),
			},
		},
	)

	if err != nil {
		return "", err
	}
	// 该url是支付二维码
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	// 对账 主动查询订单状态
	txn, _, err := n.svc.QueryOrderByOutTradeNo(
		ctx,
		native.QueryOrderByOutTradeNoRequest{
			OutTradeNo: core.String(bizTradeNO),
			Mchid:      core.String(n.mchId),
		},
	)

	if err != nil {
		return err
	}

	// 更新订单状态
	return n.updateByTxn(ctx, txn)
}

// FindExpiredPayment 过期订单
func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

// HandleCallback 支付回调也是更新订单状态
func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	return n.updateByTxn(ctx, txn)
}

func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	//
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是 %s", errUnknownTransactionsState, *txn.TradeState)
	}

	// 更新本地对应订单的状态
	err := n.repo.UpdatePayment(
		ctx,
		domain.Payment{
			// 微信传的 transaction id
			TxnID:      *txn.TransactionId,
			BizTradeNO: *txn.OutTradeNo,
			Status:     status,
		},
	)
	if err != nil {
		return err
	}
	// 通知具体的业务方
	// 有些人的系统 会根据支付状态来决定要不要通知
	// 我要是发消息失败了怎么办？
	// 站在业务的角度，你是不是至少应该发成功一次

	err = n.producer.ProducePaymentEvent(
		ctx,
		events.PaymentEvent{
			BizTradeNO: *txn.OutTradeNo,
			Status:     status.AsUint8(),
		},
	)
	if err != nil {
		n.l.Error(
			"发送支付事件失败",
			logger.Error(err),
			logger.String("biz_trade_no", *txn.OutTradeNo),
		)
	}
	return nil
}
