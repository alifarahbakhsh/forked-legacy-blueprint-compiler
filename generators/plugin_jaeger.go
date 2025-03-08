package generators

import (
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/deploy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type JaegerNode struct {
	Name            string
	TypeName        string
	Params          []Parameter
	ClientModifiers []Modifier
	ServerModifiers []Modifier
	ASTNodes        []*ServiceImplInfo
	DepInfo         *deploy.DeployInfo
}

func (n *JaegerNode) Accept(v Visitor) {
	v.VisitJaegerNode(v, n)
}

func (n *JaegerNode) GetNodes(nodeType string) []Node {
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

func GenerateJaegerNode(node parser.DetailNode) Node {
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

	return &JaegerNode{Name: node.Name, TypeName: "JaegerTracer", Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo()}
}

func (n *JaegerNode) getConstructorBody(info *parser.ImplInfo) string {
	body := ""
	body += "addr := os.Getenv(\"" + n.Name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + n.Name + "_PORT\")\n"
	body += "int_tracer := tracer.New" + n.TypeName + "(addr, port)\n"
	body += "return &" + n.Name + "{internal: int_tracer}\n"
	return body
}

func (n *JaegerNode) GenerateClientNode(info *parser.ImplInfo) {
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/choices/tracer"}, parser.ImportInfo{ImportName: "", FullName: "os"}}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "go.opentelemetry.io/otel/trace"})
	fields := []parser.ArgInfo{parser.GetPointerArg("internal", "tracer."+n.TypeName)}
	bodies := make(map[string]string)
	bodies[con_name] = n.getConstructorBody(info)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		bodies[name] = "return t.internal." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}
	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: "t", Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, PluginName: "Jaeger"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
