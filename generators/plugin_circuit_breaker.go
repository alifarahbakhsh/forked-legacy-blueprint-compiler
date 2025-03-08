package generators

import (
	"fmt"
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type CircuitBreakerModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *CircuitBreakerModifier) Accept(v Visitor) {
	v.VisitCircuitBreakerModifier(v, m)
}

func (m *CircuitBreakerModifier) GetParams() []Parameter {
	return m.Params
}

func (n *CircuitBreakerModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *CircuitBreakerModifier) GetName() string {
	return "CircuitBreakerModifier"
}

func (m *CircuitBreakerModifier) generateClientMethodBody(receiver_name string, finfo parser.FuncInfo) string {
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
	body += "if !" + receiver_name + ".cb.Ready() {\n"
	body += "\t" + ret_names[len(ret_names)-1] + " = errors.New(\"Circuit breaker tripped\")\n"
	body += "\treturn " + strings.Join(ret_names, ", ") + "\n"
	body += "}\n"
	body += "defer func(){" + receiver_name + ".cb.Done(ctx, " + ret_names[len(ret_names)-1] + ")}()\n"
	body += strings.Join(ret_names, ",") + " = " + receiver_name + ".client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	body += "return " + strings.Join(ret_names, ", ")
	return body
}

func (m *CircuitBreakerModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "time"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "log"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/mercari/go-circuitbreaker"})
	return imports
}

func (m *CircuitBreakerModifier) getClientFields(next_node_name string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("client", next_node_name))
	fields = append(fields, parser.GetPointerArg("cb", "circuitbreaker.CircuitBreaker"))
	return fields
}

func (m *CircuitBreakerModifier) getClientConstructor(name string, next_node *ServiceImplInfo) (parser.FuncInfo, string) {
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
	body += "interval_dur, err := time.ParseDuration(interval)\n"
	body += "if err != nil {\n"
	body += "\tlog.Fatal(err)\n"
	body += "}\n"
	body += "cb := circuitbreaker.New(\n"
	body += "\tcircuitbreaker.WithFailOnContextCancel(true),\n"
	body += "\tcircuitbreaker.WithFailOnContextDeadline(true),\n"
	body += "\tcircuitbreaker.WithCounterResetInterval(interval_dur),\n"
	body += "\tcircuitbreaker.WithTripFunc(circuitbreaker.NewTripFuncFailureRate(1000,0.1)),\n"
	body += ")\n"
	body += "return &" + name + "{client:client, cb:cb}\n"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *CircuitBreakerModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "cbm"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{}
	name := prev_node.BaseName + "CircuitBreaker"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

func (m *CircuitBreakerModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func (m *CircuitBreakerModifier) GetPluginName() string {
	return "CircuitBreaker"
}

func GenerateCircuitBreakerModifier(node parser.ModifierNode) Modifier {
	return &CircuitBreakerModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}
