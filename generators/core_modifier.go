package generators

import (
	"log"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/deploy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type ModifierRegistry struct {
	Registry   map[string]func(parser.ModifierNode) Modifier
	Boundaries map[string]bool
	Starts     map[string]bool
	logger     *log.Logger
}

func InitModifierRegistry(logger *log.Logger) *ModifierRegistry {
	reg := make(map[string]func(parser.ModifierNode) Modifier)
	boundaries := make(map[string]bool)
	starts := make(map[string]bool)

	// This is where you add Modifier generating functions to the registry
	reg["TracerModifier"] = GenerateTracerModifier
	reg["RPCServer"] = GenerateRPCServerModifier
	reg["ClientPool"] = GenerateClientPoolModifier
	reg["MetricModifier"] = GenerateMetricModifier
	reg["XTraceModifier"] = GenerateXTraceModifier
	reg["WebServer"] = GenerateWebServerModifier
	reg["PlatformReplication"] = GeneratePlaformReplicationModifier
	reg["Retry"] = GenerateRetryModifier
	reg["LoadBalancer"] = GenerateLoadBalancerModifier
	reg["CircuitBreaker"] = GenerateCircuitBreakerModifier
	reg["HealthChecker"] = GenerateHealthCheckModifier
	reg["ConsulModifier"] = GenerateConsulModifier

	// Modifiers that are at network boundaries
	boundaries["RPCServer"] = true
	boundaries["WebServer"] = true

	// Modifiers that are at the server-client boundaries
	starts["ClientPool"] = true

	return &ModifierRegistry{Registry: reg, Boundaries: boundaries, Starts: starts, logger: logger}
}

func get_params(node parser.ModifierNode) []Parameter {
	var params []Parameter
	for _, arg := range node.ModifierParams {
		params = append(params, convert_argument_node(arg))
	}
	return params
}

func (r *ModifierRegistry) GetModifier(node parser.ModifierNode) Modifier {
	if fn, ok := r.Registry[node.ModifierType]; ok {
		return fn(node)
	}

	r.logger.Fatal("No registered Modifier generator found for:", node.ModifierType)

	return nil
}

func (r *ModifierRegistry) OrderModifiers(modifiers []Modifier) []Modifier {
	var nwo_modifiers []Modifier
	var lo_modifiers []Modifier
	var fo_modifiers []Modifier
	for _, modifier := range modifiers {
		name := modifier.GetName()
		if _, ok := r.Boundaries[name]; ok {
			lo_modifiers = append(lo_modifiers, modifier)
		} else if _, ok2 := r.Starts[name]; ok2 {
			fo_modifiers = append(fo_modifiers, modifier)
		} else {
			nwo_modifiers = append(nwo_modifiers, modifier)
		}
	}
	nwo_modifiers = append(nwo_modifiers, lo_modifiers...)
	fo_modifiers = append(fo_modifiers, nwo_modifiers...)
	return fo_modifiers
}

type DefaultModifier struct {
}

func (m *DefaultModifier) Accept(v Visitor) {
	// Do Nothing: Don't need to visit anything
}

func (m *DefaultModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(m) == nodeType {
		nodes = append(nodes, m)
	}
	return nodes
}

func (m *DefaultModifier) GetName() string {
	return "DefaultModifier"
}

func (m *DefaultModifier) IsSourceCodeModifier() bool {
	return false
}

func (m *DefaultModifier) IsDeployerModifier() bool {
	return false
}

func (m *DefaultModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	return nil, nil
}

func (m *DefaultModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	return nil, nil
}

func (m *DefaultModifier) ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	return nil, nil
}

func (m *DefaultModifier) ModifyDeployInfo(depInfo *deploy.DeployInfo) error {
	return nil
}

func (m *DefaultModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {}

func (m *DefaultModifier) GetParams() []Parameter {
	return []Parameter{}
}

type NoOpSourceCodeModifier struct {
	*DefaultModifier
}

func (m *NoOpSourceCodeModifier) IsSourceCodeModifier() bool {
	return true
}

func NewNoOpSourceCodeModifier() *NoOpSourceCodeModifier {
	return &NoOpSourceCodeModifier{&DefaultModifier{}}
}

type NoOpDeployerModifier struct {
	*DefaultModifier
}

func (m *NoOpDeployerModifier) IsDeployerModifier() bool {
	return true
}

func NewNoOpDeployerModifier() *NoOpDeployerModifier {
	return &NoOpDeployerModifier{&DefaultModifier{}}
}
