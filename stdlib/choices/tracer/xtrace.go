package tracer

import (
	"context"
	"github.com/tracingplane/tracingplane-go/tracingplane"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/xtrace/client"
	"gitlab.mpi-sws.org/cld/tracing/tracing-framework-go/localbaggage"
	"log"
)

type XTracerImpl struct {
}

func NewXTracerImpl(addr string, port string) *XTracerImpl {
	err := client.Connect(addr + ":" + port)
	if err != nil {
		log.Println(err)
	}
	return &XTracerImpl{}
}

func (xt *XTracerImpl) Log(ctx context.Context, msg string) context.Context {
	return client.Log(ctx, msg)
}

func (xt *XTracerImpl) LogWithTags(ctx context.Context, msg string, tags ...string) context.Context {
	return client.LogWithTags(ctx, msg, tags...)
}

func (xt *XTracerImpl) StartTask(ctx context.Context, tags ...string) context.Context {
	return client.StartTask(ctx, tags...)
}

func (xt *XTracerImpl) StopTask(ctx context.Context) context.Context {
	return client.StopTask(ctx)
}

func (xt *XTracerImpl) Merge(ctx context.Context, other tracingplane.BaggageContext) context.Context {
	return localbaggage.Merge(ctx, other)
}

func (xt *XTracerImpl) Set(ctx context.Context, baggage tracingplane.BaggageContext) context.Context {
	return localbaggage.Set(ctx, baggage)
}

func (xt *XTracerImpl) Get(ctx context.Context) tracingplane.BaggageContext {
	return localbaggage.Get(ctx)
}

func (xt *XTracerImpl) IsTracing(ctx context.Context) bool {
	return client.HasTask(ctx)
}