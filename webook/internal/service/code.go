package service

import (
	"context"
	"fmt"
	"github.com/Anwenya/GeekTime/webook/internal/repository"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeVerifyTooMany

type CodeService struct {
	repo *repository.CodeRepository
	sms  sms.SMService
}

func NewCodeService(repo *repository.CodeRepository, smsSvc sms.SMService) *CodeService {
	return &CodeService{
		repo: repo,
		sms:  smsSvc,
	}
}

func (svc *CodeService) Send(ctx context.Context, biz, phone string) error {
	// 生成验证码
	code := svc.generate()
	// 存储到redis中 同时能解决并发问题
	err := svc.repo.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 真正发送验证码
	const codeTplId = "953406"
	return svc.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (svc *CodeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	// 在redis中验证
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	// 对外面屏蔽了验证次数过多的错误
	if err == repository.ErrCodeVerifyTooMany {
		return false, nil
	}
	return ok, err
}

func (svc *CodeService) generate() string {
	// 0-999999
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
