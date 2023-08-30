// Blueprint: auto-generated by Tracer plugin
package proc2

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib/components"
	"context"
	"go.opentelemetry.io/otel/trace"
)

type NonLeafServiceImplTracer struct {
	service *NonLeafServiceImpl
	tracer components.Tracer
	service_name string
}
func NewNonLeafServiceImplTracer(service *NonLeafServiceImpl,tracer components.Tracer,service_name string,sampling_rate string) *NonLeafServiceImplTracer {
	return &NonLeafServiceImplTracer{service: service, tracer: tracer, service_name: service_name}
}

func (t *NonLeafServiceImplTracer) Leaf(ctx context.Context, a int64, jaegerTracer_trace_ctx string) (int64, error) {
	if jaegerTracer_trace_ctx != "" {
		span_ctx_config, _ := components.GetSpanContext(jaegerTracer_trace_ctx)
		span_ctx := trace.NewSpanContext(span_ctx_config)
		ctx = trace.ContextWithRemoteSpanContext(ctx, span_ctx)
	}
	tp, _ := t.tracer.GetTracerProvider()
	tr := tp.Tracer(t.service_name)
	ctx, span := tr.Start(ctx, "Leaf")
	defer span.End()
	ret0, err := t.service.Leaf(ctx, a)
	if err != nil {
		span.RecordError(err)
	}
	return ret0, err
}

