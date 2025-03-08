package generators

import (
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/deploy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type RabbitMQNode struct {
	Name            string
	TypeName        string
	queue_name      string
	Params          []Parameter
	ClientModifiers []Modifier
	ServerModifiers []Modifier
	ASTNodes        []*ServiceImplInfo
	DepInfo         *deploy.DeployInfo
}

func (n *RabbitMQNode) Accept(v Visitor) {
	v.VisitRabbitMQNode(v, n)
}

func (n *RabbitMQNode) GetNodes(nodeType string) []Node {
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

func GenerateRabbitMQNode(node parser.DetailNode) Node {
	var params []Parameter
	var cmodifiers []Modifier
	var smodifiers []Modifier
	var queue_name string
	for _, arg := range node.Arguments {
		param := convert_argument_node(arg)
		switch ptype := param.(type) {
		case *ValueParameter:
			if ptype.KeywordName == "queue_name" {
				queue_name = ptype.Value
			}
		}
		params = append(params, param)
	}
	for _, modifier := range node.ClientModifiers {
		cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
	}
	for _, modifier := range node.ServerModifiers {
		smodifiers = append(smodifiers, convert_modifier_node(modifier))
	}

	return &RabbitMQNode{Name: node.Name, TypeName: "RabbitMQ", queue_name: queue_name, Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo()}
}

func (n *RabbitMQNode) getConstructorBody(info *parser.ImplInfo) string {
	body := ""
	body += "addr := os.Getenv(\"" + n.Name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + n.Name + "_PORT\")\n"
	body += "queue_name := " + strings.ReplaceAll(n.queue_name, `'`, `"`) + "\n"
	body += "int_queue := queue.NewRabbitMQ(queue_name, addr, port)\n"
	body += "return &" + n.Name + "{internal: int_queue}\n"
	return body
}

func (n *RabbitMQNode) GenerateDefaultNode(info *parser.ImplInfo) {
	methods := copyMap(info.Methods)
	bodies := make(map[string]string)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		bodies[name] = "return c.client." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}
	client_node := &ServiceImplInfo{Name: n.Name + "Default", ReceiverName: "c", Methods: methods, MethodBodies: bodies, BaseName: n.Name}
	n.ASTNodes = append(n.ASTNodes, client_node)
}

func (n *RabbitMQNode) GenerateClientNode(info *parser.ImplInfo) {
	n.GenerateDefaultNode(info)
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/choices/queue"}, parser.ImportInfo{ImportName: "", FullName: "os"}}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/components"})
	fields := []parser.ArgInfo{parser.GetPointerArg("internal", "nosqldb."+n.TypeName)}
	bodies := make(map[string]string)
	bodies[con_name] = n.getConstructorBody(info)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		bodies[name] = "return c.internal." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}
	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: "c", Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, PluginName: "RabbitMQ"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
