package generators

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type RetryModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *RetryModifier) Accept(v Visitor) {
	v.VisitRetryModifier(v, m)
}

func (m *RetryModifier) GetParams() []Parameter {
	return m.Params
}

func (n *RetryModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *RetryModifier) GetName() string {
	return "RetryModifier"
}

func (m *RetryModifier) GetPluginName() string {
	return "Retry"
}

func (m *RetryModifier) generateClientMethodBody(receiverName string, finfo parser.FuncInfo) string {
	var body string
	var arg_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	var ret_names []string
	// Add ret_names and vars to the body
	for idx, ret := range finfo.Return {
		ret_name := fmt.Sprintf("ret_%d", idx)
		ret_names = append(ret_names, ret_name)
		body += "var " + ret_name + " " + ret.String() + "\n"
	}
	body += "var i int64\n"
	body += "for i = 0; i < " + receiverName + ".max_retries; i++ {\n"
	body += "\t" + strings.Join(ret_names, ",") + " = " + receiverName + ".client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	body += "\tif " + ret_names[len(ret_names)-1] + " == nil {\n"
	body += "\t\tbreak\n"
	body += "\t}\n"
	body += "}\n"
	body += "return " + strings.Join(ret_names, ", ")
	return body
}

func (m *RetryModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "strconv"})
	return imports
}

func (m *RetryModifier) getClientFields(next_node_name string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("client", next_node_name))
	fields = append(fields, parser.GetBasicArg("max_retries", "int64"))
	return fields
}

func (m *RetryModifier) getClientConstructor(name string, next_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{}
	args = append(args, parser.GetPointerArg("client", next_node.Name))
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		}
	}
	body := ""
	body += "max_retries_num, _ := strconv.ParseInt(max_retries, 10, 64)\n"
	body += "return &" + name + "{client:client, max_retries: max_retries_num}\n"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *RetryModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "rm"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{}
	name := prev_node.BaseName + "Retrier"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

func (m *RetryModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func GenerateRetryModifier(node parser.ModifierNode) Modifier {
	return &RetryModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}
