package grpc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/api/proto/gen/payment"
	"github.com/Anwenya/GeekTime/webook/payment/domain"
	"github.com/Anwenya/GeekTime/webook/payment/service/wechat"
	"google.golang.org/grpc"
)

type WechatServiceServer struct {
	pmtv1.UnimplementedWechatPaymentServiceServer
	svc *wechat.NativePaymentService
}

func NewWechatServiceServer(svc *wechat.NativePaymentService) *WechatServiceServer {
	return &WechatServiceServer{svc: svc}
}

// NativePrePay 发起支付请求
func (w *WechatServiceServer) NativePrePay(ctx context.Context, request *pmtv1.PrePayRequest) (*pmtv1.NativePrePayResponse, error) {
	codeURL, err := w.svc.Prepay(
		ctx,
		domain.Payment{
			Amt: domain.Amount{
				Currency: request.Amt.Currency,
				Total:    request.Amt.Total,
			},
			BizTradeNO:  request.BizTradeNo,
			Description: request.Description,
		},
	)
	if err != nil {
		return nil, err
	}

	return &pmtv1.NativePrePayResponse{
		CodeUrl: codeURL,
	}, nil
}

// GetPayment 查询支付信息
func (w *WechatServiceServer) GetPayment(ctx context.Context, request *pmtv1.GetPaymentRequest) (*pmtv1.GetPaymentResponse, error) {
	p, err := w.svc.GetPayment(ctx, request.GetBizTradeNo())
	if err != nil {
		return nil, err
	}
	return &pmtv1.GetPaymentResponse{
		Status: pmtv1.PaymentStatus(p.Status),
	}, nil
}

func (w *WechatServiceServer) Register(server *grpc.Server) {
	pmtv1.RegisterWechatPaymentServiceServer(server, w)
}
