package generators

import (
	"fmt"
	"log"
	"strings"
)

type PrintVisitor struct {
	DefaultVisitor
	logger      *log.Logger
	indentLevel int
	printString string
}

func NewPrintVisitor(logger *log.Logger) *PrintVisitor {
	return &PrintVisitor{DefaultVisitor{}, logger, 0, ""}
}

func (v *PrintVisitor) getIndentString() string {
	return strings.Repeat("\t", v.indentLevel)
}

func (v *PrintVisitor) Print() {
	v.logger.Println(v.printString)
}

func (v *PrintVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.printString += "\nSystem:\n"
	v.DefaultVisitor.VisitMillenialNode(v, n)
}

func (v *PrintVisitor) VisitAnsibleContainerNode(_ Visitor, n *AnsibleContainerNode) {
	v.printString += n.Name + " = AnsibleContainer {\n"
	v.indentLevel += 1
	v.DefaultVisitor.VisitAnsibleContainerNode(v, n)
	v.indentLevel -= 1
	v.printString += "}\n"
}

func (v *PrintVisitor) VisitKubernetesContainerNode(_ Visitor, n *KubernetesContainerNode) {
	v.printString += n.Name + " = KubernetesContainer {\n"
	v.indentLevel += 1
	v.DefaultVisitor.VisitKubernetesContainerNode(v, n)
	v.indentLevel -= 1
	v.printString += "}\n"
}

func (v *PrintVisitor) VisitDockerContainerNode(_ Visitor, n *DockerContainerNode) {
	v.printString += n.Name + " = DockerContainer {\n"
	v.indentLevel += 1
	v.DefaultVisitor.VisitDockerContainerNode(v, n)
	v.printString += v.getIndentString() + "Options = " + fmt.Sprintf("%v", n.Options) + "\n"
	v.indentLevel -= 1
	v.printString += "}\n"
}

func (v *PrintVisitor) VisitProcessNode(_ Visitor, n *ProcessNode) {
	v.printString += v.getIndentString() + n.Name + " = Process {\n"
	v.indentLevel += 1
	v.DefaultVisitor.VisitProcessNode(v, n)
	v.indentLevel -= 1
	v.printString += v.getIndentString() + "}\n"
}

func (v *PrintVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.printString += v.getIndentString() + n.Name + " = " + n.Type + "("
	for idx, param := range n.Params {
		v.indentLevel += 1
		param.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.Params)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithClient("
	for idx, modifier := range n.ClientModifiers {
		v.indentLevel += 1
		modifier.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.ClientModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithServer("
	for idx, modifier := range n.ServerModifiers {
		v.indentLevel += 1
		modifier.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.ServerModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ")\n"
}

func (v *PrintVisitor) VisitQueueServiceNode(_ Visitor, n *QueueServiceNode) {
	v.printString += v.getIndentString() + n.Name + " = " + n.Type + "("
	for idx, param := range n.Params {
		v.indentLevel += 1
		param.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.Params)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithClient("
	for idx, modifier := range n.ClientModifiers {
		v.indentLevel += 1
		modifier.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.ClientModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithServer("
	for idx, modifier := range n.ServerModifiers {
		v.indentLevel += 1
		modifier.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.ServerModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ")\n"
}

func (v *PrintVisitor) VisitInstanceParameter(_ Visitor, n *InstanceParameter) {
	v.printString += n.KeywordName + "=" + n.Name + ".With("
	for idx, modifier := range n.ClientModifiers {
		v.indentLevel += 1
		modifier.Accept(v)
		v.indentLevel -= 1
		if idx != len(n.ClientModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ")"
}

func (v *PrintVisitor) VisitValueParameter(_ Visitor, n *ValueParameter) {
	v.printString += n.KeywordName + "=" + n.Value
}

func (v *PrintVisitor) modifier_str(prefix string, modifier_name string, params []Parameter) {
	v.printString += modifier_name + "("
	for idx, param := range params {
		param.Accept(v)
		if idx != len(params)-1 {
			v.printString += ","
		}
	}
	v.printString += ")"
}

func (v *PrintVisitor) VisitTracerModifier(_ Visitor, n *TracerModifier) {
	v.modifier_str(v.getIndentString(), "TracerModifier", n.Params)
}

func (v *PrintVisitor) VisitRPCServerModifier(_ Visitor, n *RPCServerModifier) {
	v.modifier_str(v.getIndentString(), "RPCServerModifier", n.Params)
}

func (v *PrintVisitor) VisitWebServerModifier(_ Visitor, n *WebServerModifier) {
	v.modifier_str(v.getIndentString(), "WebServerModifier", n.Params)
}

func (v *PrintVisitor) VisitClientPoolModifier(_ Visitor, n *ClientPoolModifier) {
	v.modifier_str(v.getIndentString(), "ClientPool", n.Params)
}

func (v *PrintVisitor) VisitMetricModifier(_ Visitor, n *MetricModifier) {
	v.modifier_str(v.getIndentString(), "MetricModifier", n.Params)
}

func (v *PrintVisitor) VisitXTraceModifier(_ Visitor, n *XTraceModifier) {
	v.modifier_str(v.getIndentString(), "XTraceModifier", n.Params)
}

func (v *PrintVisitor) VisitPlaformReplicationModifier(_ Visitor, n *PlatformReplicationModifier) {
	v.modifier_str(v.getIndentString(), "PlatformReplicationModifier", n.Params)
}

func (v *PrintVisitor) VisitRetryModifier(_ Visitor, n *RetryModifier) {
	v.modifier_str(v.getIndentString(), "RetryModifier", n.Params)
}

func (v *PrintVisitor) VisitLoadBalancerModifier(_ Visitor, n *LoadBalancerModifier) {
	v.modifier_str(v.getIndentString(), "LoadBalancerModifier", n.Params)
}

func (v *PrintVisitor) VisitCircuitBreakerModifier(_ Visitor, n *CircuitBreakerModifier) {
	v.modifier_str(v.getIndentString(), "CircuitBreakerModifier", n.Params)
}

func (v *PrintVisitor) VisitHealthCheckModifier(_ Visitor, n *HealthCheckModifier) {
	v.modifier_str(v.getIndentString(), "HealthCheckModifier", n.Params)
}

func (v *PrintVisitor) VisitConsulModifier(_ Visitor, n *ConsulModifier) {
	v.modifier_str(v.getIndentString(), "ConsulModifier", n.Params)
}

func (v *PrintVisitor) component_str(name string, node_name string, params []Parameter, ClientModifiers []Modifier, ServerModifiers []Modifier) {
	v.printString += v.getIndentString() + name + " = " + node_name + "("
	for idx, param := range params {
		param.Accept(v)
		if idx != len(params)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithClient("
	for idx, modifier := range ClientModifiers {
		modifier.Accept(v)
		if idx != len(ClientModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ").WithServer("
	for idx, modifier := range ServerModifiers {
		modifier.Accept(v)
		if idx != len(ServerModifiers)-1 {
			v.printString += ","
		}
	}
	v.printString += ")\n"
}

func (v *PrintVisitor) VisitLoadBalancerNode(_ Visitor, n *LoadBalancerNode) {
	v.component_str(n.Name, "LoadBalancerNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitJaegerNode(_ Visitor, n *JaegerNode) {
	v.component_str(n.Name, "JaegerNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitZipkinNode(_ Visitor, n *ZipkinNode) {
	v.component_str(n.Name, "ZipkinNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitLocalMetricNode(_ Visitor, n *LocalMetricNode) {
	v.component_str(n.Name, "LocalMetricNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitXTraceNode(_ Visitor, n *XTraceNode) {
	v.component_str(n.Name, "XTraceServerNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitMemcachedNode(_ Visitor, n *MemcachedNode) {
	v.component_str(n.Name, "MemcachedNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitRedisNode(_ Visitor, n *RedisNode) {
	v.component_str(n.Name, "RedisNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitMongoDBNode(_ Visitor, n *MongoDBNode) {
	v.component_str(n.Name, "MongoDBNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitRabbitMQNode(_ Visitor, n *RabbitMQNode) {
	v.component_str(n.Name, "RabbitMQNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitMySqlDBNode(_ Visitor, n *MySqlDBNode) {
	v.component_str(n.Name, "MySqlDBNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}

func (v *PrintVisitor) VisitConsulNode(_ Visitor, n *ConsulNode) {
	v.component_str(n.Name, "ConsulNode", n.Params, n.ClientModifiers, n.ServerModifiers)
}
