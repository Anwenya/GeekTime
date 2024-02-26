package tencent

import (
	"context"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
)

type Service struct {
	client   *sms.Client
	appId    *string
	signName *string
}

func (s *Service) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	request := sms.NewSendSmsRequest()
	request.SetContext(ctx)
	// 短信服务id
	request.SmsSdkAppId = s.appId
	// 短信服务签名
	request.SignName = s.signName
	// 短信模板id
	request.TemplateId = common.StringPtr(tplId)
	// 短信模板参数
	request.TemplateParamSet = common.StringPtrs(args)
	request.PhoneNumberSet = common.StringPtrs(numbers)
	response, err := s.client.SendSms(request)

	// 处理异常
	if err != nil {
		return err
	}

	for _, statusPtr := range response.Response.SendStatusSet {
		if statusPtr == nil {
			continue
		}
		// 不为nil 解引用
		status := *statusPtr
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败 code: %s, msg: %s", *status.Code, *status.Message)
		}

	}
	return nil
}

func NewService(client *sms.Client, appId string, signName string) *Service {
	return &Service{
		client:   client,
		appId:    &appId,
		signName: &signName,
	}
}
