package generators

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type LoadBalancerModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *LoadBalancerModifier) Accept(v Visitor) {
	v.VisitLoadBalancerModifier(v, m)
}

func (m *LoadBalancerModifier) GetParams() []Parameter {
	return m.Params
}

func (n *LoadBalancerModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *LoadBalancerModifier) GetName() string {
	return "LoadBalancer"
}

func (m *LoadBalancerModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	return imports
}

func (m *LoadBalancerModifier) generateClientMethodBody(receiverName string, finfo parser.FuncInfo) string {
	var arg_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	body := ""
	body += "client := " + receiverName + ".balancer.PickClient()\n"
	body += "return client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")"
	return body
}

func (m *LoadBalancerModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "lb"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{}
	name := prev_node.BaseName + "LoadBalancer"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

func (m *LoadBalancerModifier) getClientConstructor(name string, next_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{}
	args = append(args, parser.GetListArg("clients", next_node.BaseName))
	body := ""
	body += "lb := stdlib.NewLoadBalancer[" + next_node.BaseName + "](clients)\n"
	body += "return &" + name + "{lb:lb}\n"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *LoadBalancerModifier) getClientFields(next_node_name string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("balancer", "stdlib.LoadBalancer[*"+next_node_name+"]"))
	return fields
}

func (m *LoadBalancerModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func (m *LoadBalancerModifier) GetPluginName() string {
	return "LoadBalancer"
}

func GenerateLoadBalancerModifier(node parser.ModifierNode) Modifier {
	return &LoadBalancerModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}

type LoadBalancerNode struct {
	Name                string
	TypeName            string
	Params              []Parameter
	ClientModifiers     []Modifier
	ServerModifiers     []Modifier
	ASTNodes            []*ServiceImplInfo
	DepInfo             *deploy.DeployInfo
	BaseTypeName        string
	ParamClientNodes    map[string][]*ServiceImplInfo
	ModifierClientNodes map[string][]*ServiceImplInfo
}

func (n *LoadBalancerNode) Accept(v Visitor) {
	v.VisitLoadBalancerNode(v, n)
}

func (n *LoadBalancerNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	for _, child := range n.ClientModifiers {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	for _, child := range n.ServerModifiers {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func GenerateLoadBalancerNode(node parser.DetailNode) Node {
	var params []Parameter
	var cmodifiers []Modifier
	var smodifiers []Modifier
	var bTypeName string
	for _, arg := range node.Arguments {
		param := convert_argument_node(arg)
		params = append(params, param)
		if param.GetKeywordName() == "basetype" {
			bTypeName = param.GetValue()
		}
	}
	for _, modifier := range node.ClientModifiers {
		cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
	}
	for _, modifier := range node.ServerModifiers {
		smodifiers = append(smodifiers, convert_modifier_node(modifier))
	}

	return &LoadBalancerNode{Name: node.Name, TypeName: "LoadBalancer", Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo(), BaseTypeName: bTypeName, ParamClientNodes: make(map[string][]*ServiceImplInfo), ModifierClientNodes: make(map[string][]*ServiceImplInfo)}
}

func (n *LoadBalancerNode) GetDependencies() map[string]Dependency {
	dependencies := make(map[string]Dependency)
	for _, param := range n.Params {
		switch pt := param.(type) {
		case *InstanceParameter:
			dependencies[pt.Name] = Dependency{InstanceName: pt.Name, ClientModifiers: pt.ClientModifiers}
		case *ValueParameter:
			if pt.KeywordName == "clients" {
				value := strings.ReplaceAll(pt.Value, "[", "")
				value = strings.ReplaceAll(value, "]", "")
				dependency_strings := strings.Split(value, ", ")
				for _, d := range dependency_strings {
					dependencies[d] = Dependency{InstanceName: d}
				}
				continue
			}
			continue
		}
	}
	return dependencies
}

func (n *LoadBalancerNode) GenerateClientNode(info *parser.ServiceInfo) {
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_args = append(con_args, parser.GetVariadicArg("clients", "services."+n.BaseTypeName))
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: "genz/stdlib"}, parser.ImportInfo{ImportName: "", FullName: "context"}, parser.ImportInfo{ImportName: "", FullName: "spec/services"}}
	fields := []parser.ArgInfo{parser.GetPointerArg("balancer", "stdlib.LoadBalancer[services."+n.BaseTypeName+"]")}

	cons_body := ""
	cons_body += "lb := stdlib.NewLoadBalancer[services." + n.BaseTypeName + "](clients)\n"
	cons_body += "return &" + n.Name + "{balancer:lb}\n"
	bodies := make(map[string]string)
	bodies[con_name] = cons_body
	receiverName := "lb"
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		body := ""
		body += "client := " + receiverName + ".balancer.PickClient()\n"
		body += "return client." + name + "(" + strings.Join(arg_names, ", ") + ")"
		bodies[name] = body
	}
	var values []string
	deps := n.GetDependencies()
	for d, _ := range deps {
		values = append(values, d)
	}
	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: receiverName, Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, Values: values, BaseName: n.BaseTypeName, PluginName: "LoadBalancer"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
