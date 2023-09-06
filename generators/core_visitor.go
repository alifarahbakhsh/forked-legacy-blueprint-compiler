package generators

type Visitor interface {
	// Root Node
	VisitMillenialNode(v Visitor, n *MillenialNode)

	// Container Nodes
	VisitAnsibleContainerNode(v Visitor, n *AnsibleContainerNode)
	VisitDockerContainerNode(v Visitor, n *DockerContainerNode)
	VisitKubernetesContainerNode(v Visitor, n *KubernetesContainerNode)
	VisitNoOpContainerNode(v Visitor, n *NoOpContainerNode)

	// Process Nodes
	VisitProcessNode(v Visitor, n *ProcessNode)

	// Service Nodes
	VisitFuncServiceNode(v Visitor, n *FuncServiceNode)
	VisitQueueServiceNode(v Visitor, n *QueueServiceNode)

	// Parameters
	VisitInstanceParameter(v Visitor, n *InstanceParameter)
	VisitValueParameter(v Visitor, n *ValueParameter)

	// Modifiers
	VisitTracerModifier(v Visitor, n *TracerModifier)
	VisitRPCServerModifier(v Visitor, n *RPCServerModifier)
	VisitWebServerModifier(v Visitor, n *WebServerModifier)
	VisitClientPoolModifier(v Visitor, n *ClientPoolModifier)
	VisitMetricModifier(v Visitor, n *MetricModifier)
	VisitXTraceModifier(v Visitor, n *XTraceModifier)
	VisitPlaformReplicationModifier(v Visitor, n *PlatformReplicationModifier)
	VisitRetryModifier(v Visitor, n *RetryModifier)
	VisitLoadBalancerModifier(v Visitor, n *LoadBalancerModifier)
	VisitCircuitBreakerModifier(v Visitor, n *CircuitBreakerModifier)
	VisitHealthCheckModifier(v Visitor, n *HealthCheckModifier)
	VisitConsulModifier(v Visitor, n *ConsulModifier)

	// Component Nodes
	VisitLoadBalancerNode(v Visitor, n *LoadBalancerNode)

	// Choice Nodes
	VisitJaegerNode(v Visitor, n *JaegerNode)
	VisitZipkinNode(v Visitor, n *ZipkinNode)
	VisitLocalMetricNode(v Visitor, n *LocalMetricNode)
	VisitXTraceNode(v Visitor, n *XTraceNode)
	VisitMemcachedNode(v Visitor, n *MemcachedNode)
	VisitRedisNode(v Visitor, n *RedisNode)
	VisitMongoDBNode(v Visitor, n *MongoDBNode)
	VisitRabbitMQNode(v Visitor, n *RabbitMQNode)
	VisitMySqlDBNode(v Visitor, n *MySqlDBNode)
	VisitConsulNode(v Visitor, n *ConsulNode)
}

type DefaultVisitor struct{}

func (_ *DefaultVisitor) VisitMillenialNode(v Visitor, n *MillenialNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitAnsibleContainerNode(v Visitor, n *AnsibleContainerNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitKubernetesContainerNode(v Visitor, n *KubernetesContainerNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitDockerContainerNode(v Visitor, n *DockerContainerNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitNoOpContainerNode(v Visitor, n *NoOpContainerNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitProcessNode(v Visitor, n *ProcessNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitFuncServiceNode(v Visitor, n *FuncServiceNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitQueueServiceNode(v Visitor, n *QueueServiceNode) {
	for _, node := range n.Children {
		node.Accept(v)
	}
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitValueParameter(v Visitor, n *ValueParameter) {
	return
}

func (_ *DefaultVisitor) VisitInstanceParameter(v Visitor, n *InstanceParameter) {
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitTracerModifier(v Visitor, n *TracerModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitRPCServerModifier(v Visitor, n *RPCServerModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitWebServerModifier(v Visitor, n *WebServerModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitMetricModifier(v Visitor, n *MetricModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitClientPoolModifier(v Visitor, n *ClientPoolModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitXTraceModifier(v Visitor, n *XTraceModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitPlaformReplicationModifier(v Visitor, n *PlatformReplicationModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitRetryModifier(v Visitor, n *RetryModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitLoadBalancerModifier(v Visitor, n *LoadBalancerModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitCircuitBreakerModifier(v Visitor, n *CircuitBreakerModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitHealthCheckModifier(v Visitor, n *HealthCheckModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitConsulModifier(v Visitor, n *ConsulModifier) {
	for _, node := range n.Params {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitLoadBalancerNode(v Visitor, n *LoadBalancerNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitLocalMetricNode(v Visitor, n *LocalMetricNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitJaegerNode(v Visitor, n *JaegerNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitZipkinNode(v Visitor, n *ZipkinNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitXTraceNode(v Visitor, n *XTraceNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitMemcachedNode(v Visitor, n *MemcachedNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitRedisNode(v Visitor, n *RedisNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitMongoDBNode(v Visitor, n *MongoDBNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitRabbitMQNode(v Visitor, n *RabbitMQNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitMySqlDBNode(v Visitor, n *MySqlDBNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}

func (_ *DefaultVisitor) VisitConsulNode(v Visitor, n *ConsulNode) {
	for _, node := range n.Params {
		node.Accept(v)
	}
	for _, node := range n.ClientModifiers {
		node.Accept(v)
	}
	for _, node := range n.ServerModifiers {
		node.Accept(v)
	}
}
