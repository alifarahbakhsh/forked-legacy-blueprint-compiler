package tracer

import (
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/exporters/jaeger"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"log"
)

type JaegerTracer struct {
	tp *tracesdk.TracerProvider
}

func NewJaegerTracer(addr string, port string) *JaegerTracer {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint("http://" + addr + ":" + port + "/api/traces")))
	if err != nil {
		log.Fatal(err)
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
	)
	return &JaegerTracer{tp}
}

func (t * JaegerTracer) GetTracerProvider() (trace.TracerProvider, error) {
	return t.tp, nil
}
