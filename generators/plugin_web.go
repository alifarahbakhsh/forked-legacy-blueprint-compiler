package generators

import (
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/netgen"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type WebServerModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *WebServerModifier) Accept(v Visitor) {
	v.VisitWebServerModifier(v, m)
}

func (m *WebServerModifier) GetParams() []Parameter {
	return m.Params
}

func (n *WebServerModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func GenerateWebServerModifier(node parser.ModifierNode) Modifier {
	return &WebServerModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}

func (m *WebServerModifier) GetName() string {
	return "WebServer"
}

func (m *WebServerModifier) GetPluginName() string {
	return "WebServer"
}

func (m *WebServerModifier) getFrameworkName() (string, error) {
	for _, p := range m.Params {
		switch param := p.(type) {
		case *InstanceParameter:
			continue
		case *ValueParameter:
			if param.KeywordName == "framework" {
				return param.Value, nil
			}
		}
	}
	return "", errors.New("Framework not found")
}

func (m *WebServerModifier) isMetricsOn() bool {
	for _, p := range m.Params {
		switch param := p.(type) {
		case *InstanceParameter:
			continue
		case *ValueParameter:
			if param.KeywordName == "metrics" {
				if param.Value == "True" {
					return true
				}
				return false
			}
		}
	}
	return false
}

func (m *WebServerModifier) getTimeout() string {
	timeout := ""
	for _, p := range m.Params {
		switch param := p.(type) {
		case *InstanceParameter:
			continue
		case *ValueParameter:
			if param.KeywordName == "timeout" {
				timeout = param.Value
			}
		}
	}
	return timeout
}

func (m *WebServerModifier) GetFrameworkInfo() (string, netgen.NetworkGenerator, error) {
	framework, err := m.getFrameworkName()
	if err != nil {
		return "", nil, err
	}
	factory := netgen.GetNetGenFactory()
	generator, err := factory.GetGenerator(framework)
	if err != nil {
		return "", nil, err
	}
	return framework, generator, err
}

func (m *WebServerModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	_, generator, err := m.GetFrameworkInfo()
	if err != nil {
		return nil, err
	}
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "webhandler"
	name := prev_node.BaseName + "Handler"
	is_metrics_on := m.isMetricsOn()
	bodies, err := generator.GenerateServerMethods(receiver_name, prev_node.BaseName, newMethods, is_metrics_on, prev_node.InstanceName)
	if err != nil {
		return nil, err
	}
	imports := generator.GetImports(prev_node.HasUserDefinedObjs || prev_node.HasReturnDefinedObjs)
	constructor, body, cons_imports, fields, structs := generator.GenerateServerConstructor(prev_node.Name, prev_node.InstanceName, name, prev_node.BaseName, is_metrics_on)
	for _, p := range m.Params {
		switch param := p.(type) {
		case *InstanceParameter:
			continue
		case *ValueParameter:
			constructor.Args = append(constructor.Args, parser.GetBasicArg(param.KeywordName, "string"))
		}
	}
	bodies[constructor.Name] = body
	imports = append(imports, cons_imports...)
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, InstanceName: prev_node.InstanceName, Structs: structs, BaseImports: prev_node.BaseImports}, nil
}

func (m *WebServerModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	_, generator, err := m.GetFrameworkInfo()
	if err != nil {
		return nil, err
	}
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "webclient"
	name := prev_node.BaseName + "WebClient"
	is_metrics_on := m.isMetricsOn()
	has_timeout := (m.getTimeout() != "")
	bodies, err := generator.GenerateClientMethods(receiver_name, prev_node.BaseName, newMethods, prev_node.NextNodeMethodArgs, prev_node.NextNodeMethodReturn, is_metrics_on, has_timeout)
	if err != nil {
		return nil, err
	}
	imports := generator.GetImports(prev_node.HasUserDefinedObjs || prev_node.HasReturnDefinedObjs)
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: imports, InstanceName: prev_node.InstanceName}, nil
}

func (m *WebServerModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	_, generator, _ := m.GetFrameworkInfo()
	is_metrics_on := m.isMetricsOn()
	constructor, body, cons_imports, fields, structs := generator.GenerateClientConstructor(node.InstanceName, node.Name, node.BaseName, is_metrics_on, m.getTimeout())
	node.MethodBodies[constructor.Name] = body
	node.Imports = append(node.Imports, cons_imports...)
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = fields
	node.Structs = structs
}
