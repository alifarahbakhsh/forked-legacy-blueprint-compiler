package netgen

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type ThriftGenerator struct {
	appName           string
	remoteTypes       map[string]RemoteTypeInfo
	serviceTypes      map[string]ServiceInfo
	responseTypes     map[string]ResponseInfo // response_name -> response_infos
	service_responses map[string]string       //service_name + ":" + func_name -> response_name
	enums             map[string]bool
	functions         map[string][]string // service_name -> list of functions
}

func NewThriftGenerator() NetworkGenerator {
	return &ThriftGenerator{appName: "", remoteTypes: make(map[string]RemoteTypeInfo), serviceTypes: make(map[string]ServiceInfo), responseTypes: make(map[string]ResponseInfo), service_responses: make(map[string]string), enums: make(map[string]bool), functions: make(map[string][]string)}
}

func (t *ThriftGenerator) SetAppName(appName string) {
	// Can only be set once
	if t.appName == "" {
		t.appName = appName
	}
}

func (t *ThriftGenerator) GetRequirements() []parser.RequireInfo {
	return []parser.RequireInfo{parser.RequireInfo{Name: "gen-go/" + t.appName, Path: "../gen-go/" + t.appName, Version: "v1.0.0"}}
}

func (t *ThriftGenerator) GetImports(hasUserDefinedObjs bool) []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "gen-go/" + t.appName})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/apache/thrift/lib/go/thrift"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	if hasUserDefinedObjs {
		imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/jinzhu/copier"})
	}
	return imports
}

func (t *ThriftGenerator) GenerateFiles(outdir string) error {
	// Generate thrift file
	thrift_file := path.Join(outdir, t.appName+".thrift")
	f, err := os.Create(thrift_file)
	if err != nil {
		return err
	}
	var rtype_string string
	for _, rtype := range t.remoteTypes {
		rtype_string += rtype.Val + "\n"
	}
	_, err = f.WriteString(rtype_string + "\n")
	if err != nil {
		return err
	}
	var restype_string string
	for _, restype := range t.responseTypes {
		restype_string += restype.Val + "\n"
	}
	_, err = f.WriteString(restype_string + "\n")
	if err != nil {
		return err
	}
	for name, mtype := range t.serviceTypes {
		serviceString := "service " + name + "{\n"
		for _, method := range mtype.Methods {
			serviceString += "\t" + strings.ReplaceAll(method.Val, "\n", "\n\t") + "\n"
		}
		serviceString += "}\n"
		_, err := f.WriteString(serviceString + "\n")
		if err != nil {
			return err
		}
	}
	// Execute the thrift file
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(outdir)
	cmd := exec.Command("thrift", "--gen", "go", t.appName+".thrift")
	out, err := cmd.Output()
	log.Println("Generating thrift code")
	log.Println("\n" + string(out))
	if err != nil {
		return err
	}
	os.Chdir(wd)
	gen_dir := path.Join(outdir, "gen-go", strings.ToLower(t.appName))
	mod_file := path.Join(gen_dir, "go.mod")
	outf, err := os.OpenFile(mod_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer outf.Close()
	_, err = outf.WriteString("module " + t.appName + "\n\ngo 1.18\n\nrequire github.com/apache/thrift v0.16.0")
	if err != nil {
		return err
	}
	os.Chdir(gen_dir)
	e_cmd := exec.Command("go", "mod", "tidy")
	output, err := e_cmd.CombinedOutput()
	if err != nil {
		log.Println(string(output))
		return err
	}
	os.Chdir(wd)
	return nil
}

func (t *ThriftGenerator) ConvertRemoteTypes(remoteTypes map[string]*parser.ImplInfo) error {
	for name, rtype := range remoteTypes {
		var fields []FieldInfo
		rtype_string := "struct " + name + " {\n"
		for idx, field := range rtype.Fields {
			str, err := t.getThriftTypeString(field.Type)
			if err != nil {
				return err
			}
			rtype_string += "\t" + fmt.Sprintf("%d", idx+1) + ": " + str + " " + field.Name + ";\n"
			fields = append(fields, FieldInfo{Name: field.Name, Type: field.Type})
		}
		rtype_string += "}"
		rinfo := RemoteTypeInfo{Name: name, Val: rtype_string, Fields: fields}
		t.remoteTypes[name] = rinfo
	}
	return nil
}

func (t *ThriftGenerator) ConvertEnumTypes(enumTypes map[string]*parser.EnumInfo) error {
	for name, etype := range enumTypes {
		if len(etype.ValNames) == 0 {
			continue
		}
		etype_string := "enum " + name + " {\n"
		for idx, valname := range etype.ValNames {
			etype_string += "\t" + valname + " = " + fmt.Sprintf("%d", idx)
			if idx != len(etype.ValNames)-1 {
				etype_string += ","
			}
			etype_string += "\n"
		}
		etype_string += "}"
		rinfo := RemoteTypeInfo{Name: name, Val: etype_string, IsEnum: true}
		t.remoteTypes[name] = rinfo
		t.enums[name] = true
	}
	return nil
}

func (t *ThriftGenerator) generateServiceMethod(handler_name string, service_name string, funcInfo parser.FuncInfo, is_metrics_on bool) (string, error) {
	var body string
	if is_metrics_on {
		body += generateFunctionWrapperBody(handler_name, funcInfo.Name)
	}
	var argNames []string
	for idx, arg := range funcInfo.Args {
		if arg.Type.BaseType == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			// TODO: Update body to have Userdefined objects
			argType, err := t.getThriftTypeString(arg.Type)
			if err != nil {
				return "", err
			}
			if _, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + argName + " " + arg.Type.Detail.UserType + "\n"
				body += "copier.Copy(&" + argName + ", " + arg.Name + ")\n"
			}
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := t.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			if err != nil {
				return "", err
			}
			if _, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + argName + " []" + arg.Type.Detail.UserType + "\n"
				body += "copier.Copy(&" + argName + ", " + arg.Name + ")\n"
			}
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			argName := fmt.Sprint("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := t.basicTypeToString(arg.Type.ContainerType2, arg.Type.Container2Detail)
			if err != nil {
				return "", err
			}
			if _, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += argName + " := map[" + arg.Type.Detail.String(false) + "]" + arg.Type.Detail.UserType + "\n"
				body += "copier.Copy(&" + argName + ", " + arg.Name + ")\n"
			}
		} else {
			argNames = append(argNames, arg.Name)
		}
	}
	var retNames []string
	for idx, _ := range funcInfo.Return {
		retNames = append(retNames, fmt.Sprintf("ret%d", idx))
	}
	body += strings.Join(retNames, ",") + " := " + handler_name + ".service." + funcInfo.Name + "(" + strings.Join(argNames, ",") + ")\n"
	for idx, arg := range funcInfo.Return {
		if arg.Type.BaseType == parser.USERDEFINED {
			// Update body to have Userdefined objects
			new_retname := fmt.Sprintf("ret_updated%d", idx)
			argType, err := t.getThriftTypeString(arg.Type)
			if err != nil {
				return "", nil
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				if v.IsEnum {
					body += new_retname + " := " + v.Name + "(" + retNames[idx] + ")\n"
				} else {
					body += new_retname + " := " + t.appName + ".New" + v.Name + "()" + "\n"
					body += "copier.Copy(" + new_retname + ", &" + retNames[idx] + ")\n"
				}
			}
			retNames[idx] = new_retname
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			new_retname := fmt.Sprintf("ret_updated%d", idx)
			argType, err := t.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			if err != nil {
				return "", nil
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + new_retname + " []*" + t.appName + "." + v.Name + "\n"
				body += "copier.Copy(&" + new_retname + ", &" + retNames[idx] + ")\n"
			}
			retNames[idx] = new_retname
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			new_retname := fmt.Sprintf("ret_updated%d", idx)
			argType, err := t.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			if err != nil {
				return "", nil
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += new_retname + " := map[" + arg.Type.Detail.String(false) + "]*" + t.appName + "." + v.Name + "{}\n"
				body += "copier.Copy(&" + new_retname + ", &" + retNames[idx] + ")\n"
			}
			retNames[idx] = new_retname
		}
	}
	response_name := t.service_responses[service_name+":"+funcInfo.Name]
	if response_type, ok := t.responseTypes[response_name]; !ok {
		return "", errors.New("Failed to find response type for service " + service_name + " function " + funcInfo.Name)
	} else {
		new_retname := "response"
		body += new_retname + " := " + t.appName + ".New" + response_type.Name + "()\n"
		for idx, field := range response_type.Fields {
			body += new_retname + "." + field.Name + " = " + retNames[idx] + "\n"
		}
	}
	body += "return response," + retNames[len(retNames)-1]
	return body, nil
}

func (t *ThriftGenerator) getThriftArgName(name string) string {
	splits := strings.Split(name, ".")
	new_name := splits[len(splits)-1]
	if v, ok := t.remoteTypes[new_name]; ok {
		return v.Name
	}

	log.Fatal("Could not find user-defined Type ", name, " in Thrift Converted types")
	return ""
}

func (t *ThriftGenerator) packResponse(service_name string, func_name string, retVals []parser.ArgInfo) (string, error) {
	if len(retVals) == 1 {
		if _, ok := t.responseTypes["BaseRPCResponse"]; !ok {
			t.responseTypes["BaseRPCResponse"] = ResponseInfo{Name: "BaseRPCResponse", Val: "struct BaseRPCResponse {\n}"}
		}
		t.service_responses[service_name+":"+func_name] = "BaseRPCResponse"
		return "BaseRPCResponse", nil
	}
	response_name := service_name + "_" + func_name + "Response"
	response_string := "struct " + response_name + "{\n"
	var fields []FieldInfo
	for idx, retarg := range retVals {
		if idx != len(retVals)-1 {
			retType, err := t.getThriftTypeString(retarg.Type)
			if err != nil {
				return "", err
			}
			retName := retarg.Name
			if retName == "" {
				retName = fmt.Sprintf("RetVal%d", idx)
			}
			response_string += "\t" + fmt.Sprintf("%d: ", idx+1) + retType + " " + retName + ";\n"
			fields = append(fields, FieldInfo{Name: retName, Type: retarg.Type})
		}
	}
	response_string += "}\n"
	t.service_responses[service_name+":"+func_name] = response_name
	t.responseTypes[response_name] = ResponseInfo{Name: response_name, Val: response_string, Fields: fields}
	return response_name, nil
}

func (t *ThriftGenerator) GenerateServerConstructor(prev_handler string, service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetErrorArg("")}
	args := []parser.ArgInfo{parser.GetPointerArg("old_handler", prev_handler)}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	fields := []parser.ArgInfo{parser.GetPointerArg("service", prev_handler)}
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "os"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	body := ""
	body += "var protocolFactory thrift.TProtocolFactory\n"
	body += "protocolFactory = thrift.NewTBinaryProtocolFactory(true, true)\n"
	body += "var transportFactory thrift.TTransportFactory\n"
	body += "transportFactory = thrift.NewTTransportFactory()\n"
	body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn errors.New(\"Address or Port were not set\")\n}\n"
	body += "var transport thrift.TServerTransport\n"
	body += "var err error\n"
	body += "transport, err = thrift.NewTServerSocket(addr + \":\" + port)\n"
	body += "if err != nil {\n\treturn err\n}\n"
	body += "handler := &" + handler_name + "{service:old_handler}\n"
	body += "processor := " + t.appName + ".New" + base_name + "Processor(handler)\n"
	body += "server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)\n"
	if is_metrics_on {
		funcs := t.functions[base_name]
		fields = append(fields, generateMetricFields(funcs)...)
		imports = append(imports, generateMetricImports()...)
		body += generateMetricConstructorBody("handler")
	}
	body += "return server.Serve()"
	return funcInfo, body, imports, fields, []parser.StructInfo{}
}

func (t *ThriftGenerator) GenerateClientConstructor(service_name string, handler_name string, base_name string, is_metrics_on bool, timeout string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", handler_name), parser.GetErrorArg("")}
	args := []parser.ArgInfo{}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	var imports []parser.ImportInfo
	fields := []parser.ArgInfo{parser.GetPointerArg("client", t.appName+"."+base_name+"Client")}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "os"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	body := ""
	body += "var transportFactory thrift.TTransportFactory\n"
	body += "transportFactory = thrift.NewTTransportFactory()\n"
	body += "var protocolFactory thrift.TProtocolFactory\n"
	body += "protocolFactory = thrift.NewTBinaryProtocolFactory(true, true)\n"
	body += "var transport thrift.TTransport\n"
	body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn nil, errors.New(\"Address or port were not set\")\n}\n"
	body += "var err error\n"
	transport_str := "transport, err = thrift.NewTSocket(addr + \":\" + port)\n"
	retBody := "return &" + base_name + "RPCClient{client:" + t.appName + ".New" + base_name + "Client(thrift.NewTStandardClient(iprot, oprot))}, nil"
	if timeout != "" {
		tmp := "duration, err := time.ParseDuration(\"" + timeout + "\")\n"
		tmp += "if err != nil {\n"
		tmp += "\t" + transport_str
		tmp += "} else {\n"
		tmp += "\ttransport, err = thrift.NewTSocketTimeout(addr + \":\" + port, duration, duration)\n"
		tmp += "}\n"
		transport_str = tmp
		imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "time"})
		fields = append(fields, parser.GetBasicArg("Timeout", "time.Duration"))
		retBody = "return &" + base_name + "RPCClient{client:" + t.appName + ".New" + base_name + "Client(thrift.NewTStandardClient(iprot, oprot)), Timeout: duration}, nil"
	}
	body += transport_str
	body += "if err != nil {\n\treturn nil, err\n}\n"
	body += "transport, err = transportFactory.GetTransport(transport)\n"
	body += "if err != nil {\n\treturn nil, err\n}\n"
	body += "err = transport.Open()\n"
	body += "if err != nil {\n\treturn nil, err\n}\n"
	body += "iprot := protocolFactory.GetProtocol(transport)\n"
	body += "oprot := protocolFactory.GetProtocol(transport)\n"
	body += retBody
	return funcInfo, body, imports, fields, []parser.StructInfo{}
}

func (t *ThriftGenerator) GenerateServerMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, is_metrics_on bool) (map[string]string, error) {
	bodies := make(map[string]string)
	var methodInfos []MethodInfo
	var funcNames []string
	for name, method := range methods {
		// Pack return types into a single response object
		response_name, err := t.packResponse(service_name, name, method.Return)
		if err != nil {
			return bodies, err
		}
		body, err := t.generateServiceMethod(handler_name, service_name, method, is_metrics_on)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
		var new_args []parser.ArgInfo
		var new_rets []parser.ArgInfo
		// Arguments need to be modified. If it is a userdefined arg then it needs to be changed into a thrift arg
		var method_string string
		method_string += response_name + " " + name + "(\n"
		for idx, arg := range method.Args {
			if idx != 0 {
				// Only ignore the context arg which will always be the 1st arg
				arg_type, err := t.getThriftTypeString(arg.Type)
				if err != nil {
					return bodies, err
				}
				method_string += "\t" + fmt.Sprintf("%d: ", idx) + arg_type + " " + arg.Name + ";\n"
			}
			new_arg := arg

			if arg.Type.BaseType == parser.USERDEFINED {
				arg_name := t.getThriftArgName(arg.Type.Detail.UserType)
				thrift_arg_name := t.appName + "." + arg_name
				is_enum := false
				if _, ok := t.enums[arg_name]; ok {
					is_enum = true
				}
				if !is_enum {
					new_arg = parser.GetPointerArg(arg.Name, thrift_arg_name)
				} else {
					new_arg = parser.GetBasicArg(arg.Name, thrift_arg_name)
				}
			}
			new_args = append(new_args, new_arg)
		}
		method_string += ")\n"
		method.Args = new_args
		new_rets = append(new_rets, parser.GetPointerArg("", t.appName+"."+response_name))
		new_rets = append(new_rets, parser.GetErrorArg(""))
		method.Return = new_rets
		methods[name] = method
		methodInfos = append(methodInfos, MethodInfo{Name: name, Val: method_string})
		funcNames = append(funcNames, name)
	}
	if is_metrics_on {
		// Add a metric method called startMetrics
		method, body := generateMetricMethod(handler_name, service_name, funcNames)
		methods[method.Name] = method
		bodies[method.Name] = body
	}
	t.serviceTypes[service_name] = ServiceInfo{Name: service_name, Methods: methodInfos}
	t.functions[service_name] = funcNames
	return bodies, nil
}

func (t *ThriftGenerator) generateClientMethod(handler_name string, service_name string, funcInfo parser.FuncInfo, has_timeout bool) (string, error) {
	var body string
	var argNames []string
	for idx, arg := range funcInfo.Args {
		if idx == 0 {
			argNames = append(argNames, arg.Name)
			if has_timeout {
				body += arg.Name + ", cancel := context.WithTimeout(" + arg.Name + "," + handler_name + ".Timeout)\n"
				body += "defer cancel()\n"
			}
			continue
		}
		if arg.Type.BaseType == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := t.getThriftTypeString(arg.Type)
			if err != nil {
				return "", err
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				if v.IsEnum {
					body += "var " + argName + " " + t.appName + "." + v.Name + "\n"
					body += "copier.Copy(&" + argName + ", &" + arg.Name + ")\n"
				} else {
					body += argName + " := " + t.appName + ".New" + v.Name + "()\n"
					body += "copier.Copy(" + argName + ", &" + arg.Name + ")\n"
				}
			}
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := t.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			if err != nil {
				return "", err
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + argName + " []*" + t.appName + "." + v.Name + "{}\n"
				body += "copier.Copy(&" + argName + ", &" + arg.Name + ")\n"
			}
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := t.basicTypeToString(arg.Type.ContainerType2, arg.Type.Container2Detail)
			if err != nil {
				return "", err
			}
			if v, ok := t.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += argName + " := map[" + arg.Type.Detail.String(false) + "]*" + t.appName + "." + v.Name + "{}\n"
				body += "copier.Copy(&" + argName + ", &" + arg.Name + ")\n"
			}
		} else {
			argNames = append(argNames, arg.Name)
		}
	}
	var retNames []string
	if len(funcInfo.Return) > 1 {
		body += "response, err := " + handler_name + ".client." + funcInfo.Name + "(" + strings.Join(argNames, ",") + ")\n"
	} else {
		body += "_, err := " + handler_name + ".client." + funcInfo.Name + "(" + strings.Join(argNames, ",") + ")\n"
	}
	resp_name := t.service_responses[service_name+":"+funcInfo.Name]
	response_body := ""
	if respType, ok := t.responseTypes[resp_name]; !ok {
		return "", errors.New("Response object not found for function " + service_name + "." + funcInfo.Name)
	} else {
		for idx, field := range respType.Fields {
			if field.Type.BaseType == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := t.getThriftTypeString(field.Type)
				if err != nil {
					return "", err
				}
				if _, ok := t.remoteTypes[argType]; !ok {
					log.Println("Unknowns Userdefined type in responsetype:", argType)
					return "", errors.New("Unknown Userdefined type: " + argType)
				} else {
					body += retName + " := " + field.Type.Detail.UserType + "{}\n"
					response_body += "copier.Copy(&" + retName + ", response." + field.Name + ")\n"
				}
				retNames = append(retNames, retName)
			} else if field.Type.BaseType == parser.LIST && field.Type.ContainerType1 == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := t.basicTypeToString(field.Type.ContainerType1, field.Type.Detail)
				if err != nil {
					return "", err
				}
				if _, ok := t.remoteTypes[argType]; !ok {
					return "", errors.New("Unknown Userdefined type: " + argType)
				} else {
					body += retName + " := []" + field.Type.Detail.UserType + "{}\n"
					response_body += "copier.Copy(&" + retName + ", response." + field.Name + ")\n"
				}
				retNames = append(retNames, retName)
			} else if field.Type.BaseType == parser.MAP && field.Type.ContainerType2 == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := t.basicTypeToString(field.Type.ContainerType2, field.Type.Detail)
				if err != nil {
					return "", err
				}
				if _, ok := t.remoteTypes[argType]; !ok {
					return "", errors.New("Unknown Userdefined type: " + argType)
				} else {
					body += retName + " := map[" + field.Type.Detail.String(false) + "]" + field.Type.Container2Detail.UserType + "{}\n"
					response_body += "copier.Copy(&" + retName + ", response." + field.Name + ")\n"
				}
				retNames = append(retNames, retName)
			} else {
				retName := fmt.Sprintf("ret%d", idx)
				body += "var " + retName + " " + field.Type.String() + "\n"
				response_body += retName + " = response." + field.Name + "\n"
				retNames = append(retNames, retName)
			}
		}
	}
	retNames = append(retNames, "err")
	body += "if err != nil {\n"
	body += "\treturn " + strings.Join(retNames, ",") + "\n"
	body += "}\n"
	if has_timeout {
		body += "if " + argNames[0] + ".Err() != nil {\n"
		body += "\treturn " + strings.Join(retNames, ",") + "\n"
		body += "}\n"
	}
	body += response_body
	body += "return " + strings.Join(retNames, ",")
	return body, nil
}

func (t *ThriftGenerator) GenerateClientMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, nextNodeMethodArgs []parser.ArgInfo, nextNodeMethodReturn []parser.ArgInfo, is_metrics_on bool, has_timeout bool) (map[string]string, error) {
	bodies := make(map[string]string)
	for name, method := range methods {
		method.Args = append(method.Args, nextNodeMethodArgs...)
		last_return := method.Return[len(method.Return)-1]
		method.Return = append(method.Return[:len(method.Return)-1], nextNodeMethodReturn...)
		method.Return = append(method.Return, last_return)
		methods[name] = method
		body, err := t.generateClientMethod(handler_name, service_name, method, has_timeout)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
	}
	return bodies, nil
}

func (t *ThriftGenerator) basicTypeToString(baseType parser.Type, typeDetail parser.TypeDetail) (string, error) {
	if baseType == parser.BASIC {
		if typeDetail.TypeName == parser.INT64 {
			return "i64", nil
		} else if typeDetail.TypeName == parser.STRING {
			return "string", nil
		} else if typeDetail.TypeName == parser.DOUBLE {
			return "double", nil
		} else if typeDetail.TypeName == parser.BOOL {
			return "bool", nil
		}
	} else if baseType == parser.USERDEFINED {
		splits := strings.Split(typeDetail.UserType, ".")
		return splits[len(splits)-1], nil
	}
	return "", errors.New("Unsupported type for thrift: " + baseType.String())
}

func (t *ThriftGenerator) getThriftTypeString(typeInfo parser.TypeInfo) (string, error) {
	if typeInfo.BaseType == parser.BASIC || typeInfo.BaseType == parser.USERDEFINED {
		return t.basicTypeToString(typeInfo.BaseType, typeInfo.Detail)
	} else if typeInfo.BaseType == parser.LIST {
		basic_type, err := t.basicTypeToString(typeInfo.ContainerType1, typeInfo.Detail)
		if err != nil {
			return "", err
		}
		return "list<" + basic_type + ">", nil
	} else if typeInfo.BaseType == parser.MAP {
		basic_type1, err := t.basicTypeToString(typeInfo.ContainerType1, typeInfo.Detail)
		if err != nil {
			return "", err
		}
		basic_type2, err := t.basicTypeToString(typeInfo.ContainerType2, typeInfo.Container2Detail)
		if err != nil {
			return "", err
		}
		return "map<" + basic_type1 + "," + basic_type2 + ">", nil
	}
	return "", errors.New("Unsupported type for thrift: " + typeInfo.BaseType.String())
}
