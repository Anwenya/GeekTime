package grpc

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/account/domain"
	"github.com/Anwenya/GeekTime/webook/account/service"
	"github.com/Anwenya/GeekTime/webook/api/proto/gen/account/v1"
	"github.com/ecodeclub/ekit/slice"
	"google.golang.org/grpc"
)

type AccountServiceServer struct {
	accountv1.UnimplementedAccountServiceServer
	svc service.AccountService
}

func NewAccountServiceServer(svc service.AccountService) *AccountServiceServer {
	return &AccountServiceServer{svc: svc}
}

func (a *AccountServiceServer) Credit(
	ctx context.Context,
	request *accountv1.CreditRequest,
) (*accountv1.CreditResponse, error) {
	err := a.svc.Credit(ctx, a.toDomain(request))
	return &accountv1.CreditResponse{}, err
}

func (a *AccountServiceServer) toDomain(c *accountv1.CreditRequest) domain.Credit {
	return domain.Credit{
		Biz:   c.Biz,
		BizId: c.BizId,
		Items: slice.Map(
			c.Items,
			func(idx int, src *accountv1.CreditItem) domain.CreditItem {
				return a.itemToDomain(src)
			},
		),
	}
}

func (a *AccountServiceServer) itemToDomain(c *accountv1.CreditItem) domain.CreditItem {
	return domain.CreditItem{
		Uid:         c.Uid,
		Account:     c.Account,
		AccountType: domain.AccountType(c.AccountType),
		Amount:      c.Amount,
		Currency:    c.Currency,
	}
}

func (a *AccountServiceServer) Register(server *grpc.Server) {
	accountv1.RegisterAccountServiceServer(server, a)
}
