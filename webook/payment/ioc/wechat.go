package ioc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/payment/events"
	"github.com/Anwenya/GeekTime/webook/payment/repository"
	"github.com/Anwenya/GeekTime/webook/payment/service/wechat"
	"github.com/Anwenya/GeekTime/webook/pkg/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"os"
)

func InitWechatClient(cfg WechatConfig, l logger.LoggerV1) *core.Client {
	// 从本地读取相关的配置
	mchPrivateKey, err := utils.LoadPrivateKey(cfg.KeyPath)
	if err != nil {
		l.Error("读取微信支付配置失败", logger.Error(err))
		panic(any(err))
	}

	// 使用商户私钥等初始化 client
	ctx := context.Background()
	client, err := core.NewClient(
		ctx,
		option.WithWechatPayAutoAuthCipher(
			cfg.MchID, cfg.MchSerialNum,
			mchPrivateKey, cfg.MchKey,
		),
	)
	if err != nil {
		l.Error("初始化微信支付客户端失败", logger.Error(err))
		panic(any(err))
	}

	return client
}

func InitWechatNativeService(
	cli *core.Client,
	repo repository.PaymentRepository,
	cfg WechatConfig,
	producer events.Producer,
	l logger.LoggerV1,
) *wechat.NativePaymentService {
	return wechat.NewNativePaymentService(
		cfg.AppID,
		cfg.MchID,
		repo,
		&native.NativeApiService{
			Client: cli,
		},
		producer,
		l,
	)
}

func InitWechatNotifyHandler(cfg WechatConfig) *notify.Handler {
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(cfg.MchID)
	handler, err := notify.NewRSANotifyHandler(
		cfg.MchKey,
		verifiers.NewSHA256WithRSAVerifier(certificateVisitor),
	)
	if err != nil {
		panic(any(err))
	}
	return handler
}

func InitWechatConfig() WechatConfig {
	return WechatConfig{
		AppID:        os.Getenv("WEPAY_APP_ID"),
		MchID:        os.Getenv("WEPAY_MCH_ID"),
		MchKey:       os.Getenv("WEPAY_MCH_KEY"),
		MchSerialNum: os.Getenv("WEPAY_MCH_SERIAL_NUM"),
		CertPath:     "./config/cert/apiclient_cert.pem",
		KeyPath:      "./config/cert/apiclient_key.pem",
	}
}

type WechatConfig struct {
	AppID        string
	MchID        string
	MchKey       string
	MchSerialNum string

	// 证书
	CertPath string
	KeyPath  string
}
