package generators

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type TracerModifier struct {
	*NoOpSourceCodeModifier
	tracerInstanceName string
	Params             []Parameter
}

func (m *TracerModifier) Accept(v Visitor) {
	v.VisitTracerModifier(v, m)
}

func (m *TracerModifier) GetParams() []Parameter {
	return m.Params
}

func (n *TracerModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *TracerModifier) GetName() string {
	return "TracerModifier"
}

func (m *TracerModifier) GetPluginName() string {
	return "Tracer"
}

func (m *TracerModifier) generateServerMethodBody(receiverName string, finfo parser.FuncInfo, argname string) string {
	var arg_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	var ret_names []string
	for idx, _ := range finfo.Return {
		ret_names = append(ret_names, fmt.Sprintf("ret%d", idx))
	}
	ret_names[len(ret_names)-1] = "err"
	body := ""
	body += "if " + argname + " != \"\" {\n"
	body += "\tspan_ctx_config, _ := components.GetSpanContext(" + argname + ")\n"
	body += "\tspan_ctx := trace.NewSpanContext(span_ctx_config)\n"
	body += "\tctx = trace.ContextWithRemoteSpanContext(ctx, span_ctx)\n"
	body += "}\n"
	body += "tp, _ := " + receiverName + ".tracer.GetTracerProvider()\n"
	body += "tr := tp.Tracer(" + receiverName + ".service_name)\n"
	body += "ctx, span := tr.Start(ctx, \"" + finfo.Name + "\")\n"
	body += "defer span.End()\n"
	body += strings.Join(ret_names, ", ") + " := " + receiverName + ".service." + finfo.Name + "(" + strings.Join(arg_names[:len(arg_names)-1], ", ") + ")\n"
	body += "if err != nil {\n"
	body += "\tspan.RecordError(err)\n"
	body += "}\n"
	body += "return " + strings.Join(ret_names, ", ")
	return body
}

func (m *TracerModifier) generateClientMethodBody(receiverName string, finfo parser.FuncInfo) string {
	var arg_names []string
	for _, arg := range finfo.Args {
		arg_names = append(arg_names, arg.Name)
	}
	var ret_names []string
	for idx, _ := range finfo.Return {
		ret_names = append(ret_names, fmt.Sprintf("ret%d", idx))
	}
	ret_names[len(ret_names)-1] = "err"
	trace_ctx_arg := m.tracerInstanceName + "_trace_ctx"
	body := ""
	body += "tp, _ := " + receiverName + ".tracer.GetTracerProvider()\n"
	body += "tr := tp.Tracer(" + receiverName + ".service_name)\n"
	body += "ctx, span := tr.Start(ctx, \"" + finfo.Name + "\")\n"
	body += "defer span.End()\n"
	body += trace_ctx_arg + ",_ := span.SpanContext().MarshalJSON()\n"
	arg_names = append(arg_names, "string("+trace_ctx_arg+")")
	body += strings.Join(ret_names, ", ") + " := " + receiverName + ".client." + finfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	body += "if err != nil {\n"
	body += "\tspan.RecordError(err)\n"
	body += "}\n"
	body += "return " + strings.Join(ret_names, ", ")
	return body
}

func (m *TracerModifier) getImports() []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: MODULE_ROOT + "/stdlib/components"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	return imports
}

func (m *TracerModifier) getServerImports() []parser.ImportInfo {
	imports := m.getImports()
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "go.opentelemetry.io/otel/trace"})
	return imports
}

func (m *TracerModifier) getQueueImports() []parser.ImportInfo {
	imports := m.getImports()
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "encoding/json"})
	return imports
}

func (m *TracerModifier) getServerFields(prev_node *ServiceImplInfo) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("service", prev_node.Name))
	fields = append(fields, parser.GetBasicArg("tracer", "components.Tracer"))
	fields = append(fields, parser.GetBasicArg("service_name", "string"))
	return fields
}

func (m *TracerModifier) getClientFields(next_node string) []parser.ArgInfo {
	var fields []parser.ArgInfo
	fields = append(fields, parser.GetPointerArg("client", next_node))
	fields = append(fields, parser.GetBasicArg("tracer", "components.Tracer"))
	fields = append(fields, parser.GetBasicArg("service_name", "string"))
	return fields
}

func (m *TracerModifier) getConstructor(name string, prev_node *ServiceImplInfo) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("service", prev_node.Name)}
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		case *InstanceParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "components.Tracer"))
		}
	}
	body := ""
	body = "return &" + name + "{service: service, tracer: tracer, service_name: service_name}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *TracerModifier) getClientConstructor(name string, next_node string) (parser.FuncInfo, string) {
	func_name := "New" + name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", name)}
	args := []parser.ArgInfo{parser.GetPointerArg("client", next_node)}
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "string"))
		case *InstanceParameter:
			args = append(args, parser.GetBasicArg(ptype.KeywordName, "components.Tracer"))
		}
	}
	body := ""
	body = "return &" + name + "{client: client, tracer: tracer, service_name: service_name}"
	return parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}, body
}

func (m *TracerModifier) ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "t"
	for name, method := range newMethods {
		arg_name := m.tracerInstanceName + "_trace_ctx"
		method.Args = append(method.Args, parser.GetBasicArg(arg_name, "string"))
		bodies[name] = m.generateServerMethodBody(receiver_name, method, arg_name)
		newMethods[name] = method
	}
	name := prev_node.BaseName + "Tracer"
	constructor, body := m.getConstructor(name, prev_node)
	bodies[constructor.Name] = body
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getServerImports(), Fields: m.getServerFields(prev_node), Constructors: []parser.FuncInfo{constructor}, InstanceName: prev_node.InstanceName, BaseImports: prev_node.BaseImports}, nil
}

func (m *TracerModifier) ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
	bodies := make(map[string]string)
	newMethods := copyMap(prev_node.Methods)
	receiver_name := "t"
	arg_name := m.tracerInstanceName + "_trace_ctx"
	for name, method := range newMethods {
		combineMethodInfo(&method, prev_node)
		bodies[name] = m.generateClientMethodBody(receiver_name, method)
		newMethods[name] = method
	}
	next_node_args := []parser.ArgInfo{parser.GetBasicArg(arg_name, "string")}
	name := prev_node.BaseName + "TracerClient"

	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getImports(), InstanceName: prev_node.InstanceName, NextNodeMethodArgs: next_node_args}, nil
}

func (m *TracerModifier) getQueueBody(receiver_name string, func_name string, method parser.FuncInfo) string {
	var body string
	if func_name == "Send" {
		var arg_names []string
		for _, arg := range method.Args {
			arg_names = append(arg_names, arg.Name)
		}
		body += "tp,_ := " + receiver_name + ".tracer.GetTracerProvider()\n"
		body += "\ttr := tp.Tracer(" + receiver_name + ".service_name)\n"
		body += "ctx, span := tr.Start(ctx, \"Send\")\n"
		body += "defer span.End()\n"
		payload_arg_name := arg_names[1]
		body += "msg := make(map[string]string)\n"
		body += "span_ctx, _ := span.SpanContext().MarshalJSON()\n"
		body += "msg[\"trace_ctx\"] = string(span_ctx)\n"
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
		body += "\tspan_ctx_config, _ := components.GetSpanContext(trace_str)\n"
		body += "\tspan_ctx := trace.NewSpanContext(span_ctx_config)\n"
		body += "\tctx := context.Background()\n"
		body += "\tctx = trace.ContextWithRemoteSpanContext(ctx, span_ctx)\n"
		body += "\ttp, _ := " + receiver_name + ".tracer.GetTracerProvider()\n"
		body += "\ttr := tp.Tracer(" + receiver_name + ".service_name)\n"
		body += "\tctx, span := tr.Start(ctx, \"Recv\")\n"
		body += "\tdefer span.End()\n"
		arg_name := method.Args[0].Name
		body += "\t" + arg_name + "([]bytes(msg[\"payload\"]))\n"
		body += "}\n"
		body += receiver_name + ".client.Recv(trace_fn)\n"
	}
	return body
}

func (m *TracerModifier) ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error) {
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
	name := prev_node.BaseName + "TracerClient"
	return &ServiceImplInfo{Name: name, ReceiverName: receiver_name, Methods: newMethods, MethodBodies: bodies, BaseName: prev_node.BaseName, Imports: m.getQueueImports(), InstanceName: prev_node.InstanceName}, nil
}

func (m *TracerModifier) AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo) {
	constructor, body := m.getClientConstructor(node.Name, next_node.Name)
	node.MethodBodies[constructor.Name] = body
	node.Constructors = []parser.FuncInfo{constructor}
	node.Fields = m.getClientFields(next_node.Name)
}

func GenerateTracerModifier(node parser.ModifierNode) Modifier {
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
	return &TracerModifier{NewNoOpSourceCodeModifier(), tracerInstanceName, get_params(node)}
}
