package opentelemetry

import (
	"context"
	"github.com/Anwenya/GeekTime/webook/internal/service/sms"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Decorator struct {
	svc    sms.SMService
	tracer trace.Tracer
}

func NewDecorator(svc sms.SMService, tracer trace.Tracer) *Decorator {
	return &Decorator{svc: svc, tracer: tracer}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, numbers ...string) error {
	ctx, span := d.tracer.Start(ctx, "sms")
	defer span.End()
	span.SetAttributes(attribute.String("tpl", tplId))
	span.AddEvent("发短信")
	err := d.svc.Send(ctx, tplId, args, numbers...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
