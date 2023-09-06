// Blueprint: auto-generated by Jaeger plugin
package proc1

import (
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/stdlib/choices/tracer"
	"os"
	"go.opentelemetry.io/otel/trace"
)

type jaegerTracer struct {
	internal *tracer.JaegerTracer
}
func NewjaegerTracer() *jaegerTracer {
	addr := os.Getenv("jaegerTracer_ADDRESS")
	port := os.Getenv("jaegerTracer_PORT")
	int_tracer := tracer.NewJaegerTracer(addr, port)
	return &jaegerTracer{internal: int_tracer}
	
}

func (t *jaegerTracer) GetTracerProvider() (trace.TracerProvider, error) {
	return t.internal.GetTracerProvider()
	
}

