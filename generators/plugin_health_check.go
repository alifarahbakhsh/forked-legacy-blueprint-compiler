package generators

import (
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type HealthCheckModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *HealthCheckModifier) Accept(v Visitor) {
	v.VisitHealthCheckModifier(v, m)
}

func (m *HealthCheckModifier) GetParams() []Parameter {
	return m.Params
}

func (n *HealthCheckModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *HealthCheckModifier) GetName() string {
	return "HealthCheck"
}

func (m *HealthCheckModifier) GetPluginName() string {
	return "HealthCheck"
}

func (m *HealthCheckModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	return imports
}

func (m *HealthCheckModifier) generateServerMethodBody(receiver_name string, method parser.FuncInfo) string {
	var body string
	var arg_names []string
	for _, arg := range method.Args {
		arg_names = append(arg_names, arg.Name)
	}
	body += "return " + receiver_name + ".service." + method.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	return body
}

func (m *HealthCheckModifier) getConstructor(name string, prev_node *ServiceImplInfo) (parser.FuncInfo, string) {
	var body string
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("service", prev_node.Name)}
	body = "return &" + name + "{service: service}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *HealthCheckModifier) getFields(prev_node *ServiceImplInfo) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("service", prev_node.Name))
	return fields
}

func (m *HealthCheckModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "hc"
	for name, method := range newMethods {
		bodies[name] = m.generateServerMethodBody(receiver_name, method)
	}

	// Add a new Health method
	newMethods["Health"] = parser.FuncInfo{Name: "Health", Args: []parser.ArgInfo{parser.GetContextArg("ctx")}, Return: []parser.ArgInfo{parser.GetBasicArg("", "string"), parser.GetErrorArg("")}, Public: true}
	bodies["Health"] = "return \"Healthy\", nil"

	name := prev_node.BaseName + "HealthChecker"
	constructor, body := m.getConstructor(name, prev_node)
	bodies[constructor.Name] = body
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), Fields: m.getFields(prev_node), Constructors: []parser.FuncInfo{constructor}, InstanceName: prev_node.InstanceName, BaseImports: prev_node.BaseImports}, nil
}

func GenerateHealthCheckModifier(node parser.ModifierNode) Modifier {
	params := get_params(node)
	return &HealthCheckModifier{NewNoOpSourceCodeModifier(), params}
}
