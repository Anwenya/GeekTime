package service

import (
	"context"
	"errors"
	"fmt"
	accountv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/account/v1"
	pmtv1 "github.com/Anwenya/GeekTime/webook/api/proto/gen/payment"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/Anwenya/GeekTime/webook/reward/domain"
	"github.com/Anwenya/GeekTime/webook/reward/repository"
	"strconv"
	"strings"
)

type WechatNativeRewardService struct {
	paymentCli pmtv1.WechatPaymentServiceClient
	accountCli accountv1.AccountServiceClient
	repo       repository.RewardRepository
	l          logger.LoggerV1
}

func NewWechatNativeRewardService(
	paymentCli pmtv1.WechatPaymentServiceClient,
	accountCli accountv1.AccountServiceClient,
	repo repository.RewardRepository,
	l logger.LoggerV1,
) RewardService {
	return &WechatNativeRewardService{paymentCli: paymentCli, accountCli: accountCli, repo: repo, l: l}
}

func (w *WechatNativeRewardService) PreReward(ctx context.Context, r domain.Reward) (domain.CodeURL, error) {
	res, err := w.repo.GetCachedCodeURL(ctx, r)
	if err != nil {
		return res, nil
	}

	// 创建打赏
	r.Status = domain.RewardStatusInit
	rid, err := w.repo.CreateReward(ctx, r)
	if err != nil {
		return domain.CodeURL{}, err
	}
	// 创建支付
	pmtResp, err := w.paymentCli.NativePrePay(
		ctx,
		&pmtv1.PrePayRequest{
			Amt: &pmtv1.Amount{
				Total:    r.Amount,
				Currency: "CNY",
			},
			BizTradeNo:  fmt.Sprintf("reward-%d", rid),
			Description: fmt.Sprintf("打赏-%s", r.Target.BizName),
		},
	)
	if err != nil {
		return domain.CodeURL{}, err
	}

	cu := domain.CodeURL{
		Rid: rid,
		URL: pmtResp.CodeUrl,
	}
	// 缓存获得到的支付二维码
	err = w.repo.CachedCodeURL(ctx, cu, r)
	if err != nil {
		w.l.Error(
			"缓存二维码失败",
			logger.Error(err),
			logger.Int64("rid", rid),
		)
	}
	return cu, nil
}

func (w *WechatNativeRewardService) GetReward(ctx context.Context, rid, uid int64) (domain.Reward, error) {
	// 快路径 直接查询本地
	res, err := w.repo.GetReward(ctx, rid)
	if err != nil {
		return domain.Reward{}, err
	}

	// 确保查询的是自己的打赏
	if res.Uid != uid {
		return domain.Reward{}, errors.New("非法操作")
	}

	// 降级或限流 不走满路径
	if ctx.Value("limited") == "true" {
		return res, nil
	}

	// 如果本地状态未完成 再去查询支付服务
	if !res.Completed() {
		// 主动查询
		pmtRes, err := w.paymentCli.GetPayment(
			ctx,
			&pmtv1.GetPaymentRequest{
				BizTradeNo: w.bizTradeNO(rid),
			},
		)
		if err != nil {
			w.l.Error(
				"满路径查询支付状态失败",
				logger.Error(err),
				logger.Int64("rid", rid),
			)
			return res, nil
		}

		// 根据支付状态 更新本地状态
		switch pmtRes.Status {
		case pmtv1.PaymentStatus_PaymentStatusSuccess:
			res.Status = domain.RewardStatusPayed
		case pmtv1.PaymentStatus_PaymentStatusInit:
			res.Status = domain.RewardStatusInit
		case pmtv1.PaymentStatus_PaymentStatusRefund:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusFailed:
			res.Status = domain.RewardStatusFailed
		case pmtv1.PaymentStatus_PaymentStatusUnknown:
			w.l.Warn("未知支付状态")
		}
		err = w.UpdateReward(ctx, w.bizTradeNO(rid), res.Status)
		if err != nil {
			w.l.Error(
				"慢路径更新本地状态失败",
				logger.Error(err),
				logger.Int64("rid", rid),
			)
		}

	}
	return res, nil
}

func (w *WechatNativeRewardService) UpdateReward(ctx context.Context, bizTradeNO string, status domain.RewardStatus) error {
	rid := w.toRid(bizTradeNO)
	err := w.repo.UpdateStatus(ctx, rid, status)
	if err != nil {
		return err
	}
	// 完成支付 准备入账
	if status == domain.RewardStatusPayed {
		r, err := w.repo.GetReward(ctx, rid)
		if err != nil {
			return err
		}

		// 抽成
		amt := int64(float64(r.Amount) * 0.1)

		_, err = w.accountCli.Credit(
			ctx,
			&accountv1.CreditRequest{
				Biz:   "reward",
				BizId: rid,
				Items: []*accountv1.CreditItem{
					// 平台抽成
					{
						AccountType: accountv1.AccountType_AccountTypeReward,
						Amount:      amt,
					},
					// 被打赏者入账
					{
						Account:     r.Uid,
						Uid:         r.Uid,
						AccountType: accountv1.AccountType_AccountTypeReward,
						Amount:      r.Amount - amt,
						Currency:    "CNY",
					},
				},
			},
		)

		if err != nil {
			w.l.Error(
				"入账失败 需要修复数据",
				logger.String("biz_trade_no", bizTradeNO),
				logger.Error(err),
			)
		}
	}
	return nil
}

func (w *WechatNativeRewardService) bizTradeNO(rid int64) string {
	return fmt.Sprintf("reward-%d", rid)
}

func (w *WechatNativeRewardService) toRid(tradeNO string) int64 {
	ridStr := strings.Split(tradeNO, "-")
	val, _ := strconv.ParseInt(ridStr[1], 10, 64)
	return val
}
