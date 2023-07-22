package generators

import (
	"log"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type IRExtensionRegistry struct {
	Registry map[string]func(parser.DetailNode) Node
	logger   *log.Logger
}

func InitIRRegistry(logger *log.Logger) *IRExtensionRegistry {
	reg := make(map[string]func(parser.DetailNode) Node)

	// This is where you add Node generating functions to the registry
	reg["JaegerTracer"] = GenerateJaegerNode
	reg["ZipkinTracer"] = GenerateZipkinNode
	reg["LocalMetricCollector"] = GenerateLocalMetricNode
	reg["XTracerImpl"] = GenerateXTraceNode
	reg["Memcached"] = GenerateMemcachedNode
	reg["RedisCache"] = GenerateRedisNode
	reg["MongoDB"] = GenerateMongoDBNode
	reg["RabbitMQ"] = GenerateRabbitMQNode
	reg["MySqlDB"] = GenerateMySqlDBNode
	reg["LoadBalancer"] = GenerateLoadBalancerNode

	return &IRExtensionRegistry{Registry: reg, logger: logger}
}

func (r *IRExtensionRegistry) GetNode(node parser.DetailNode) Node {
	if fn, ok := r.Registry[node.Type]; ok {
		return fn(node)
	}

	r.logger.Fatal("No registered IR Node generator found for type:", node.Type)
	return nil
}
