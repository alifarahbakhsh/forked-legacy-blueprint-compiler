module container5

go 1.18

require (
	gen-go/leaf v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/jinzhu/copier v0.3.5
	github.com/alifarahbakhsh/forked-legacy-blueprint-compiler v0.0.1
	go.opentelemetry.io/otel/trace v1.6.0
	google.golang.org/grpc v1.41.0
	spec v1.0.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/openzipkin/zipkin-go v0.4.0 // indirect
	github.com/tracingplane/tracingplane-go v0.0.0-20171025152126-8c4e6f79b148 // indirect
	gitlab.mpi-sws.org/cld/tracing/tracing-framework-go v0.0.0-20211206181151-6edc754a9f2a // indirect
	go.opentelemetry.io/otel v1.6.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.2.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.6.0 // indirect
	go.opentelemetry.io/otel/sdk v1.6.0 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

replace gen-go/leaf v1.0.0 => ../gen-go/leaf

replace spec => ../spec
