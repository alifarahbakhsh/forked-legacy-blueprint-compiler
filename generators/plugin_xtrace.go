package generators

import (
	"fmt"
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/deploy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type XTraceModifier struct {
	*NoOpSourceCodeModifier
	tracerInstanceName string
	Params             []Parameter
}

func (m *XTraceModifier) Accept(v Visitor) {
	v.VisitXTraceModifier(v, m)
}

func (m *XTraceModifier) GetParams() []Parameter {
	return m.Params
}

func (n *XTraceModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *XTraceModifier) GetName() string {
	return "XTraceModifier"
}

func (m *XTraceModifier) GetPluginName() string {
	return "XTrace"
}

func (m *XTraceModifier) generateClientMethodBody(receiverName string, finfo parser.FuncInfo) string {
	var arg_names []string
	var ret_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	for idx, _ := range finfo.Return {
		ret_names = append(ret_names, fmt.Sprintf("ret%d", idx))
	}
	body := ""
	body += "ctx = " + receiverName + ".tracer.Log(ctx, \"" + finfo.Name + "\")\n"
	body += "baggage := " + receiverName + ".tracer.Get(ctx)\n"
	body += "baggage_str := tracingplane.EncodeBase64(baggage)\n"
	arg_names = append(arg_names, "baggage_str")
	if len(ret_names) > 1 {
		body += strings.Join(ret_names[:len(ret_names)-1], ", ") + ", "
	}
	body += "ret_baggage_str," + ret_names[len(ret_names)-1] + " := " + receiverName + ".client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	body += "ret_baggage, _ := tracingplane.DecodeBase64(ret_baggage_str)\n"
	body += "ctx = " + receiverName + ".tracer.Merge(ctx, ret_baggage)\n"
	body += "if " + ret_names[len(ret_names)-1] + " != nil {\n"
	body += "\tctx = " + receiverName + ".tracer.LogWithTags(ctx," + ret_names[len(ret_names)-1] + ".Error(), \"Error\")\n"
	for _, arg := range finfo.Args {
		body += "\tctx = " + receiverName + ".tracer.Log(ctx,\"" + arg.Name + ":\" + fmt.Sprintf(\"%v\"," + arg.Name + "))\n"
	}
	body += "}\n"
	body += "return " + strings.Join(ret_names, ", ")
	return body
}

func (m *XTraceModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/components"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/tracingplane/tracingplane-go/tracingplane"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "fmt"})
	return imports
}

func (m *XTraceModifier) getQueueImports() []parser.ImportInfo {
	imports := m.getImports()
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "encoding/json"})
	return imports
}

func (m *XTraceModifier) getClientFields(next_node string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("client", next_node))
	fields = append(fields, parser.GetBasicArg("tracer", "components.XTracer"))
	return fields
}

func (m *XTraceModifier) getClientConstructor(name string, next_node string) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("client", next_node)}
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		case *InstanceParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "components.XTracer"))
		}
	}
	body := ""
	body = "return &" + name + "{client: client, tracer: tracer}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *XTraceModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "t"
	arg_name := m.tracerInstanceName + "_baggage"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{parser.GetBasicArg(arg_name, "string")}
	next_node_rets := []parser.ArgInfo{parser.GetBasicArg("", "string")}
	name := prev_node.BaseName + "XTracerClient"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args, NextNodeMethodReturn: next_node_rets}, nil
}

func (m *XTraceModifier) generateServerMethodBody(receiverName string, finfo parser.FuncInfo, argname string) string {
	var arg_names []string
	var ret_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	for idx, _ := range finfo.Return {
		ret_names = append(ret_names, fmt.Sprintf("ret%d", idx))
	}
	body := ""
	body += "if " + argname + " != \"\"{\n"
	body += "\tremote_baggage, _ := tracingplane.DecodeBase64(" + argname + ")\n"
	body += "\tctx = " + receiverName + ".tracer.Set(ctx, remote_baggage)\n"
	body += "}\n"
	body += "if !" + receiverName + ".tracer.IsTracing(ctx) {\n"
	body += "\tctx = " + receiverName + ".tracer.StartTask(ctx, \"" + finfo.Name + "\")\n"
	body += "}\n"
	body += "ctx = " + receiverName + ".tracer.Log(ctx, \"" + finfo.Name + " start\")\n"
	body += strings.Join(ret_names, ", ") + " := " + receiverName + ".service." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	body += "ret_baggage := " + receiverName + ".tracer.Get(ctx)\n"
	body += "ret_baggage_str := tracingplane.EncodeBase64(ret_baggage)\n"
	body += "if " + ret_names[len(ret_names)-1] + " != nil {\n"
	body += "\tctx = " + receiverName + ".tracer.LogWithTags(ctx," + ret_names[len(ret_names)-1] + ".Error(), \"Error\")\n"
	for _, arg := range finfo.Args {
		body += "\tctx = " + receiverName + ".tracer.Log(ctx,\"" + arg.Name + ":\" + fmt.Sprintf(\"%v\"," + arg.Name + "))\n"
	}
	body += "}\n"
	body += "ctx = " + receiverName + ".tracer.Log(ctx, \"" + finfo.Name + " end\")\n"
	if len(ret_names) > 1 {
		body += "return " + strings.Join(ret_names[:len(ret_names)-1], ", ") + ", ret_baggage_str," + ret_names[len(ret_names)-1]
	} else {
		body += "return ret_baggage_str," + ret_names[0]
	}
	return body
}

func (m *XTraceModifier) getQueueBody(receiver_name string, func_name string, method parser.FuncInfo) string {
	var body string
	if func_name == "Send" {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		body += "ctx = " + receiver_name + ".tracer.Log(ctx, \"Send\")\n"
		payload_arg_name := arg_names[1]
		body += "msg := make(map[string]string)\n"
		body += "baggage := " + receiver_name + ".tracer.Get(ctx)\n"
		body += "msg[\"trace_ctx\"] = tracingplane.EncodeBase64(baggage)\n"
		body += "msg[\"payload\"] = string(" + payload_arg_name + ")\n"
		body += "msg_bytes, _ := json.Marshal(msg)\n"
		arg_names[1] = "msg_bytes\n"
		body += "return " + receiver_name + ".client.Send(" + strings.Join(arg_names, ", ") + ")\n"
	} else if func_name == "Recv" {
		body += "var trace_fn components.Callback_fn\n"
		body += "trace_fn = func(arg []byte){\n"
		body += "\tvar msg map[string]string\n"
		body += "\t_ = json.Unmarshal(arg, &msg)\n"
		body += "\ttrace_str := msg[\"trace_ctx\"]\n"
		body += "\tremote_baggage, _ := tracingplane.DecodeBase64(trace_str)\n"
		body += "\tctx := context.Background()\n"
		body += "\tif !" + receiver_name + ".tracer.IsTracing(ctx){\n"
		body += "\t\tctx = " + receiver_name + ".tracer.StartTask(ctx, \"Recv\")\n"
		body += "\t}\n"
		body += "\tctx = " + receiver_name + ".tracer.Log(ctx, \"Recv start\")\n"
		arg_name := method.Args[0].Name
		body += "\t" + arg_name + "([]bytes(msg[\"payload\"]))\n"
		body += "}\n"
		body += receiver_name + ".client.Recv(trace_fn)\n"
	}
	return body
}

func (m *XTraceModifier) getServerFields(prev_node *ServiceImplInfo) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("service", prev_node.Name))
	fields = append(fields, parser.GetBasicArg("tracer", "components.XTracer"))
	return fields
}

func (m *XTraceModifier) getServerImports() []parser.ImportInfo {
	imports := m.getImports()

	return imports
}

func (m *XTraceModifier) getConstructor(name string, prev_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("service", prev_node.Name)}
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		case *InstanceParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "components.XTracer"))
		}
	}
	body := ""
	body = "return &" + name + "{service: service, tracer: tracer}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *XTraceModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "t"
	for name, method := range newMethods {
		arg_name := m.tracerInstanceName + "_baggage"
		bodies[name] = m.generateServerMethodBody(receiver_name, method, arg_name)
		method.Args = append(method.Args, parser.GetBasicArg(arg_name, "string"))
		last_return := method.Return[len(method.Return)-1]
		method.Return = append(method.Return[:len(method.Return)-1], parser.GetBasicArg("", "string"))
		method.Return = append(method.Return, last_return)
		newMethods[name] = method
	}
	name := prev_node.BaseName + "XTracer"
	constructor, body := m.getConstructor(name, prev_node)
	bodies[constructor.Name] = body
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getServerImports(), Fields: m.getServerFields(prev_node), Constructors: []parser.FuncInfo{constructor}, InstanceName: prev_node.InstanceName, BaseImports: prev_node.BaseImports}, nil
}

func (m *XTraceModifier) ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
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
	name := prev_node.BaseName + "XTracerClient"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getQueueImports(), InstanceName: prev_node.InstanceName}, nil
}

func (m *XTraceModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node.Name)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func GenerateXTraceModifier(node parser.ModifierNode) Modifier {
	params := get_params(node)
	tracerInstanceName := ""
	for _, param := range params {
		switch ptype := param.(type) {
		case *InstanceParameter:
			if ptype.KeywordName == "tracer" {
				tracerInstanceName = ptype.Name
			}
		default:
			continue
		}
	}

	return &XTraceModifier{NewNoOpSourceCodeModifier(), tracerInstanceName, get_params(node)}
}

type XTraceNode struct {
	Name            string
	TypeName        string
	Params          []Parameter
	ServerModifiers []Modifier
	ClientModifiers []Modifier
	ASTNodes        []*ServiceImplInfo
	DepInfo         *deploy.DeployInfo
}

func (n *XTraceNode) Accept(v Visitor) {
	v.VisitXTraceNode(v, n)
}

func (n *XTraceNode) GetNodes(nodeType string) []Node {
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

func GenerateXTraceNode(node parser.DetailNode) Node {
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

	return &XTraceNode{Name: node.Name, TypeName: "XTracerImpl", Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, DepInfo: deploy.NewDeployInfo()}
}

func (n *XTraceNode) getConstructorBody(info *parser.ImplInfo) string {
	body := ""
	body += "addr := os.Getenv(\"" + n.Name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + n.Name + "_PORT\")\n"
	body += "int_tracer := tracer.New" + n.TypeName + "(addr, port)\n"
	body += "return &" + n.Name + "{internal: int_tracer}\n"
	return body
}

func (n *XTraceNode) GenerateClientNode(info *parser.ImplInfo) {
	methods := copyMap(info.Methods)
	con_name := "New" + n.Name
	con_args := []parser.ArgInfo{}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", n.Name)}
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/choices/tracer"}, parser.ImportInfo{ImportName: "", FullName: "os"}}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/tracingplane/tracingplane-go/tracingplane"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	fields := []parser.ArgInfo{parser.GetPointerArg("internal", "tracer."+n.TypeName)}
	bodies := make(map[string]string)
	bodies[con_name] = n.getConstructorBody(info)
	for name, method := range methods {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		if name == "StartTask" || name == "LogWithTags" {
			arg_names[len(arg_names)-1] += "..."
		}
		bodies[name] = "return t.internal." + name + "(" + strings.Join(arg_names, ", ") + ")\n"
	}

	client_node := &ServiceImplInfo{Name: n.Name, ReceiverName: "t", Methods: methods, Constructors: []parser.FuncInfo{constructor}, Imports: imports, Fields: fields, MethodBodies: bodies, PluginName: "XTrace"}
	n.ASTNodes = append(n.ASTNodes, client_node)
}
