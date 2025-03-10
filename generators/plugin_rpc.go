package generators

import (
	"errors"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/netgen"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type RPCServerModifier struct {
	*NoOpSourceCodeModifier
	Params    []Parameter
	param_map map[string]string
}

func (m *RPCServerModifier) Accept(v Visitor) {
	v.VisitRPCServerModifier(v, m)
}

func (m *RPCServerModifier) GetParams() []Parameter {
	return m.Params
}

func (n *RPCServerModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func GenerateRPCServerModifier(node parser.ModifierNode) Modifier {
	modifier := &RPCServerModifier{NewNoOpSourceCodeModifier(), get_params(node), make(map[string]string)}
	modifier.parseCustomParams()
	return modifier
}

func (m *RPCServerModifier) GetName() string {
	return "RPCServer"
}

func (m *RPCServerModifier) GetPluginName() string {
	framework, _ := m.getFrameworkName()
	return "RPC" + framework
}

func (m *RPCServerModifier) getFrameworkName() (string, error) {
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

func (m *RPCServerModifier) isMetricsOn() bool {
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

func (m *RPCServerModifier) getTimeout() string {
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

func (m *RPCServerModifier) parseCustomParams() {
	var new_params []Parameter
	for _, p := range m.Params {
		switch param := p.(type) {
		case *InstanceParameter:
			new_params = append(new_params, p)
			continue
		case *ValueParameter:
			if param.KeywordName == "resolver" {
				m.param_map[param.KeywordName] = param.Value
			} else {
				new_params = append(new_params, p)
			}
		}
	}
	m.Params = new_params
}

func (m *RPCServerModifier) GetFrameworkInfo() (string, netgen.NetworkGenerator, error) {
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

func (m *RPCServerModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	_, generator, err := m.GetFrameworkInfo()
	if err != nil {
		return nil, err
	}
	generator.SetCustomParameters(m.param_map)
	is_metrics_on := m.isMetricsOn()
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "rpchandler"
	name := prev_node.BaseName + "Handler"
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
	var baseImports []parser.ImportInfo
	if prev_node.HasUserDefinedObjs {
		baseImports = prev_node.BaseImports
	}
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, InstanceName: prev_node.InstanceName, Structs: structs, BaseImports: baseImports}, nil
}

func (m *RPCServerModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	_, generator, err := m.GetFrameworkInfo()
	if err != nil {
		return nil, err
	}
	generator.SetCustomParameters(m.param_map)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "rpcclient"
	name := prev_node.BaseName + "RPCClient"
	has_timeout := (m.getTimeout() != "")
	bodies, err := generator.GenerateClientMethods(receiver_name, prev_node.BaseName, newMethods, prev_node.NextNodeMethodArgs, prev_node.NextNodeMethodReturn, m.isMetricsOn(), has_timeout)
	if err != nil {
		return nil, err
	}
	imports := generator.GetImports(prev_node.HasUserDefinedObjs || prev_node.HasReturnDefinedObjs)
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: imports, InstanceName: prev_node.InstanceName}, nil
}

func (m *RPCServerModifier) ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	return nil, nil
}

func (m *RPCServerModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	_, generator, _ := m.GetFrameworkInfo()
	constructor, body, cons_imports, fields, structs := generator.GenerateClientConstructor(node.InstanceName, node.Name, node.BaseName, m.isMetricsOn(), m.getTimeout())
	node.MethodBodies[constructor.Name] = body
	node.Imports = append(node.Imports, cons_imports...)
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = fields
	node.Structs = structs
}
