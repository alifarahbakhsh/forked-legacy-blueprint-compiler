package tracer

import (
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"log"
)

type ZipkinTracer struct {
	tp *tracesdk.TracerProvider
}

func NewZipkinTracer(addr string, port string) *ZipkinTracer {
	exp, err := zipkin.New("http://" + addr + ":" + port + "/api/v2/spans")
	if err != nil {
		log.Fatal(err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
	)
	return &ZipkinTracer{tp}
}

func (t *ZipkinTracer) GetTracerProvider() (trace.TracerProvider, error) {
	return t.tp, nil
}
