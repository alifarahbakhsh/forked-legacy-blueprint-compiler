package generators

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type ConsulModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *ConsulModifier) Accept(v Visitor) {
	v.VisitConsulModifier(v, m)
}

func (m *ConsulModifier) GetParams() []Parameter {
	return m.Params
}

func (n *ConsulModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *ConsulModifier) GetName() string {
	return "ConsulModifier"
}

func (m *ConsulModifier) GetPluginName() string {
	return "Consul"
}

func (m *ConsulModifier) getServerFields(prev_node *ServiceImplInfo) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("service", prev_node.Name))
	fields = append(fields, parser.GetBasicArg("reg", "components.Registry"))
	fields = append(fields, parser.GetBasicArg("service_name", "string"))
	fields = append(fields, parser.GetBasicArg("service_id", "string"))
	return fields
}

func (m *ConsulModifier) getServerImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/components"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "os"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "strconv"})
	return imports
}

func (m *ConsulModifier) getConstructor(name string, prev_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("service", prev_node.Name)}
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		case *InstanceParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "components.Registry"))
		}
	}
	body := ""
	body += "addr := os.Getenv(\"" + prev_node.InstanceName + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + prev_node.InstanceName + "_PORT\")\n"
	body += "port_val, _ := strconv.ParseInt(port, 10, 64)\n"
	body += "reg.Register(service_id, service_name, addr, port_val)\n"
	body += "return &" + name + "{service:service, service_name: service_name, reg: reg}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *ConsulModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "r"
	for name, method := range newMethods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		var body string
		body = "return " + receiver_name + ".service." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
		bodies[name] = body
	}

	name := prev_node.BaseName + "Registry"
	constructor, body := m.getConstructor(name, prev_node)
	bodies[constructor.Name] = body
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getServerImports(), Fields: m.getServerFields(prev_node), Constructors: []parser.FuncInfo{constructor}, InstanceName: prev_node.InstanceName, BaseImports: prev_node.BaseImports}, nil
}

func GenerateConsulModifier(node parser.ModifierNode) Modifier {
	return &ConsulModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}

type ConsulNode struct {
	Name            string
	TypeName        string
	Params          []Parameter
	ClientModifiers []Modifier
	ServerModifiers []Modifier
	ASTNodes        []*ServiceImplInfo
	DepInfo         *deploy.DeployInfo
}

func (n *ConsulNode) Accept(v Visitor) {
	v.VisitConsulNode(v, n)
}

func (n *ConsulNode) GetNodes(nodeType string) []Node {
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
	return nodes
}

func GenerateConsulNode(node parser.DetailNode) Node {
	var params []Parameter
	var cmodifiers []Modifier
	var smodifiers []Modifier
	for _, arg := range node.Arguments {
		params = append(params, convert_argument_node(arg))
	}
	for _, modifier := range node.ClientModifiers {
		cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
	}
	for _, modifier := range node.ServerModifiers {
		smodifiers = append(smodifiers, convert_modifier_node(modifier))
	}

	return &ConsulNode{Name: node.Name, TypeName: "ConsulRegistry", Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo()}
}

func (n *ConsulNode) getConstructorBody(info *parser.ImplInfo) string {
	body := ""
	body += "addr := os.Getenv(\"" + n.Name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + n.Name + "_PORT\")\n"
	body += "port_val, _ := strconv.ParseInt(port, 10, 64)\n"
	body += "reg, err := registry.NewConsulRegistry(addr, int(port_val))\n"
	body += "if err != nil {\n\tlog.Fatal(err)\n}\n"
	body += "return &" + n.Name + "{reg: reg}\n"
	return body
}

func (n *ConsulNode) GenerateClientNode(info *parser.ImplInfo) {
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: "os"}, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/choices/registry"}, parser.ImportInfo{ImportName: "", FullName: "strconv"}, parser.ImportInfo{ImportName: "", FullName: "log"}}
	fields := []parser.ArgInfo{parser.GetPointerArg("reg", "registry.ConsulRegistry")}
	bodies := make(map[string]string)
	bodies[con_name] = n.getConstructorBody(info)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		bodies[name] = "return r.reg." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}
	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: "r", Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, PluginName: "Consul"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
