package sms

import "context"

type SMService interface {
	Send(ctx context.Context, tplId string, args []string, numbers ...string) error
}
