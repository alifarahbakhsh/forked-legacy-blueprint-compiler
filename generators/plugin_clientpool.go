package generators

import (
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type ClientPoolModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *ClientPoolModifier) Accept(v Visitor) {
	v.VisitClientPoolModifier(v, m)
}

func (m *ClientPoolModifier) GetParams() []Parameter {
	return m.Params
}

func (n *ClientPoolModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *ClientPoolModifier) GetName() string {
	return "ClientPool"
}

func (m *ClientPoolModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "strconv"})
	return imports
}

func (m *ClientPoolModifier) generateClientMethodBody(receiverName string, finfo parser.FuncInfo) string {
	var arg_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	body := ""
	body += "client := " + receiverName + ".pool.Pop()\n"
	body += "defer " + receiverName + ".pool.Push(client)\n"
	body += "return client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")"
	return body
}

func (m *ClientPoolModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "cp"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{}
	name := prev_node.BaseName + "Clientpool"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

func (m *ClientPoolModifier) getClientConstructor(name string, next_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{}
	has_metrics := false
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			if ptype.KeywordName == "metrics" {
				has_metrics = true
			}
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		}
	}
	args = append(args, parser.GetBasicArg("fn", "func()*"+next_node.Name))
	body := ""
	body += "max_clients_num, _ := strconv.ParseInt(max_clients, 10, 64)\n"
	body += "pool := stdlib.NewClientPool[*" + next_node.Name + "](max_clients_num, fn)\n"
	if has_metrics {
		body += "if metrics == \"True\" {\n"
		body += "\tpool.StartMetricsThread(service_name)\n"
		body += "}\n"
	}
	body += "return &" + name + "{pool: pool}\n"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *ClientPoolModifier) getClientFields(next_node_name string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("pool", "stdlib.ClientPool[*"+next_node_name+"]"))
	return fields
}

func (m *ClientPoolModifier) getQueueBody(receiver_name string, func_name string, method parser.FuncInfo) string {
	var body string
	if func_name == "Send" {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		body += "client := " + receiver_name + ".pool.Pop()\n"
		body += "defer " + receiver_name + ".pool.Push(client)\n"
		body += "return client.Send(" + strings.Join(arg_names, ", ") + ")\n"
	} else if func_name == "Recv" {
		body += "client := " + receiver_name + ".pool.Pop()\n"
		body += "defer func(){ go " + receiver_name + ".pool.Push(client) }()\n"
		arg_name := method.Args[0].Name
		body += "client.Recv(" + arg_name + "\n"
	}
	return body
}

func (m *ClientPoolModifier) ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := make(map[string]parser.FuncInfo)
	receiver_name := "t"
	// Only need to take care of Send and Recv.
	for name, method := range newMethods {
		if name == "Send" || name == "Recv" {
			args := make([]parser.ArgInfo, len(method.Args))
			copy(args, method.Args)
			rets := make([]parser.ArgInfo, len(method.Return))
			copy(rets, method.Return)
			newMethods[name] = parser.FuncInfo{Name: name, Args: args, Return: rets}
			body := m.getQueueBody(receiver_name, name, method)
			bodies[name] = body
		}
	}
	name := prev_node.BaseName + "Clientpool"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName}, nil
}

func (m *ClientPoolModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func (m *ClientPoolModifier) GetPluginName() string {
	return "ClientPool"
}

func GenerateClientPoolModifier(node parser.ModifierNode) Modifier {
	return &ClientPoolModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}
