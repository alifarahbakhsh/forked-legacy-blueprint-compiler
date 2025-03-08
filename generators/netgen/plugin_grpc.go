package netgen

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type GRPCGenerator struct {
	appName           string
	remoteTypes       map[string]RemoteTypeInfo
	serviceTypes      map[string]ServiceInfo
	responseTypes     map[string]ResponseInfo // response_name -> response_infos
	requestTypes      map[string]RequestInfo  // response_name -> request_infos
	service_responses map[string]string       // service_name + ":" + func_name -> response_name
	service_requests  map[string]string       // service_name + ":" + func_name -> request_name
	functions         map[string][]string     // service_name -> list of functions
	custom_params     map[string]string
}

func NewGRPCGenerator() NetworkGenerator {
	return &GRPCGenerator{appName: "", remoteTypes: make(map[string]RemoteTypeInfo), serviceTypes: make(map[string]ServiceInfo), responseTypes: make(map[string]ResponseInfo), service_responses: make(map[string]string), requestTypes: make(map[string]RequestInfo), service_requests: make(map[string]string), functions: make(map[string][]string), custom_params: make(map[string]string)}
}

func (g *GRPCGenerator) SetAppName(appName string) {
	// Can only be set once
	if g.appName == "" {
		g.appName = appName
	}
}

func (g *GRPCGenerator) GetRequirements() []parser.RequireInfo {
	return []parser.RequireInfo{parser.RequireInfo{Name: "gen-go/" + g.appName, Path: "../gen-go/" + g.appName, Version: "v1.0.0"}}
}

func (g *GRPCGenerator) GetImports(hasUserDefinedObjs bool) []parser.ImportInfo {
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "gen-go/" + g.appName})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "google.golang.org/grpc"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	if hasUserDefinedObjs {
		imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "github.com/jinzhu/copier"})
	}
	return imports
}

func (g *GRPCGenerator) GenerateFiles(outdir string) error {
	// Generate pb file
	grpc_file := path.Join(outdir, g.appName+".proto")
	f, err := os.Create(grpc_file)
	if err != nil {
		return err
	}
	head_string := "syntax=\"proto3\";\n"
	head_string += "option go_package=\"gen-go/" + g.appName + "\";\n"
	head_string += "package " + g.appName + ";\n"
	_, err = f.WriteString(head_string + "\n")
	if err != nil {
		return err
	}
	var rtype_string string
	for _, rtype := range g.remoteTypes {
		rtype_string += rtype.Val + "\n"
	}
	_, err = f.WriteString(rtype_string + "\n")
	if err != nil {
		return err
	}
	var restype_string string
	for _, restype := range g.responseTypes {
		restype_string += restype.Val + "\n"
	}
	_, err = f.WriteString(restype_string + "\n")
	if err != nil {
		return err
	}
	var reqtype_string string
	for _, reqtype := range g.requestTypes {
		reqtype_string += reqtype.Val + "\n"
	}
	_, err = f.WriteString(reqtype_string + "\n")
	if err != nil {
		return err
	}
	for name, mtype := range g.serviceTypes {
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
	// Execute the grpc file
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	os.Chdir(outdir)
	err = os.MkdirAll("gen-go/"+g.appName, os.ModePerm)
	cmd := exec.Command("protoc", "-I=./", "--go_out=./", "--go-grpc_out=./", g.appName+".proto")
	out, err := cmd.CombinedOutput()
	log.Println("Generating GRPC code")
	log.Println("\n" + string(out))
	if err != nil {
		return err
	}
	os.Chdir(wd)
	gen_dir := path.Join(outdir, "gen-go", g.appName)
	mod_file := path.Join(gen_dir, "go.mod")
	outf, err := os.OpenFile(mod_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer outf.Close()
	mod_string := "module " + g.appName + "\n\ngo 1.14\n\nrequire google.golang.org/grpc v1.36.0\nrequire github.com/golang/protobuf v1.5.2\nrequire google.golang.org/protobuf v1.26.0"
	_, err = outf.WriteString(mod_string)
	if err != nil {
		return err
	}
	os.Chdir(gen_dir)
	e_cmd := exec.Command("go", "mod", "tidy")
	output, err := e_cmd.CombinedOutput()
	if err != nil {
		log.Println(string(output))
	}
	os.Chdir(wd)
	return nil
}

func (g *GRPCGenerator) ConvertRemoteTypes(remoteTypes map[string]*parser.ImplInfo) error {
	for name, rtype := range remoteTypes {
		var fields []FieldInfo
		rtype_string := "message " + name + " {\n"
		for idx, field := range rtype.Fields {
			str, err := g.getGrpcTypeString(field.Type)
			if err != nil {
				return err
			}
			rtype_string += "\t" + str + " " + field.Name + " = " + fmt.Sprintf("%d", idx+1) + ";\n"
			fields = append(fields, FieldInfo{Name: field.Name, Type: field.Type})
		}
		rtype_string += "}"
		rinfo := RemoteTypeInfo{Name: name, Val: rtype_string, Fields: fields}
		g.remoteTypes[name] = rinfo
	}
	return nil
}

func (g *GRPCGenerator) ConvertEnumTypes(enumTypes map[string]*parser.EnumInfo) error {
	for name, etype := range enumTypes {
		if len(etype.ValNames) == 0 {
			continue
		}
		etype_string := "enum " + name + " {\n"
		for idx, valname := range etype.ValNames {
			etype_string += "\t" + valname + " = " + fmt.Sprintf("%d", idx) + ";\n"
		}
		etype_string += "}"
		rinfo := RemoteTypeInfo{Name: name, Val: etype_string, IsEnum: true, PkgPath: etype.PkgPath}
		g.remoteTypes[name] = rinfo
	}
	return nil
}

func (g *GRPCGenerator) getGrpcArgName(name string) string {
	if v, ok := g.remoteTypes[name]; ok {
		return v.Name
	}

	log.Fatal("Could not find user-defined Type in Grpc Converted types")
	return ""
}

func (g *GRPCGenerator) packResponse(service_name string, func_name string, retVals []parser.ArgInfo) (string, error) {
	if len(retVals) == 1 {
		if _, ok := g.responseTypes["BaseRPCResponse"]; !ok {
			g.responseTypes["BaseRPCResponse"] = ResponseInfo{Name: "BaseRPCResponse", Val: "message BaseRPCResponse {\n}"}
		}
		g.service_responses[service_name+":"+func_name] = "BaseRPCResponse"
		return "BaseRPCResponse", nil
	}
	response_name := service_name + "_" + func_name + "Response"
	response_string := "message " + response_name + "{\n"
	var fields []FieldInfo
	for idx, retarg := range retVals {
		if idx != len(retVals)-1 {
			retType, err := g.getGrpcTypeString(retarg.Type)
			if err != nil {
				return "", err
			}
			retName := retarg.Name
			if retName == "" {
				retName = fmt.Sprintf("RetVal%d", idx)
			}
			response_string += "\t" + retType + " " + retName + " = " + fmt.Sprintf("%d", idx+1) + ";\n"
			fields = append(fields, FieldInfo{Name: retName, Type: retarg.Type})
		}
	}
	response_string += "}\n"
	g.service_responses[service_name+":"+func_name] = response_name
	g.responseTypes[response_name] = ResponseInfo{Name: response_name, Val: response_string, Fields: fields}
	return response_name, nil
}

func (g *GRPCGenerator) packRequest(service_name string, func_name string, args []parser.ArgInfo) (string, error) {
	if len(args) == 1 {
		if _, ok := g.requestTypes["BaseRPCRequest"]; !ok {
			g.requestTypes["BaseRPCRequest"] = RequestInfo{Name: "BaseRPCRequest", Val: "message BaseRPCRequest {\n}"}
		}
		g.service_requests[service_name+":"+func_name] = "BaseRPCRequest"
		return "BaseRPCRequest", nil
	}
	request_name := service_name + "_" + func_name + "Request"
	request_string := "message " + request_name + "{\n"
	var fields []FieldInfo
	for idx, arg := range args {
		if idx != 0 {
			// Ignore the 1st argument which is the context arg
			argType, err := g.getGrpcTypeString(arg.Type)
			if err != nil {
				return "", err
			}
			argName := arg.Name
			if argName == "" {
				argName = fmt.Sprintf("Arg%d", idx)
			}
			request_string += "\t" + argType + " " + argName + " = " + fmt.Sprintf("%d", idx) + ";\n"
			fields = append(fields, FieldInfo{Name: argName, Type: arg.Type})
		}
	}
	request_string += "}\n"
	g.service_requests[service_name+":"+func_name] = request_name
	g.requestTypes[request_name] = RequestInfo{Name: request_name, Val: request_string, Fields: fields}
	return request_name, nil
}

func (g *GRPCGenerator) convertMessageFieldName(name string) string {
	pieces := strings.Split(name, "_")
	var new_name string
	for _, p := range pieces {
		new_name += strings.Title(p)
	}
	return new_name
}

func (g *GRPCGenerator) generateServiceMethod(handler_name string, service_name string, funcInfo parser.FuncInfo, request_arg_name string, is_metrics_on bool) (string, error) {
	var body string
	if is_metrics_on {
		body += generateFunctionWrapperBody(handler_name, funcInfo.Name)
	}
	var argNames []string
	for idx, arg := range funcInfo.Args {
		if idx == 0 {
			// Special handling for context arg
			argNames = append(argNames, arg.Name)
			continue
		}
		if arg.Type.BaseType == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			// TODO: Update body to have Userdefined objects
			argType, err := g.getGrpcTypeString(arg.Type)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				if v.IsEnum {
					tokens := strings.Split(v.PkgPath, "/")
					body += "var " + argName + " " + tokens[len(tokens)-1] + "." + v.Name + "\n"
					body += "copier.Copy(&" + argName + "," + request_arg_name + "." + converted_arg_name + ")\n"
				} else {
					body += argName + " := " + arg.Type.Detail.UserType + "{}\n"
					body += "copier.Copy(&" + argName + ", " + request_arg_name + "." + converted_arg_name + ")\n"
				}
			}
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := g.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if _, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += argName + " := []" + arg.Type.Detail.UserType + "{}\n"
				body += "copier.Copy(&" + argName + ", " + request_arg_name + "." + converted_arg_name + "}\n"
			}
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := g.basicTypeToString(arg.Type.ContainerType2, arg.Type.Container2Detail)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if _, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknwon Userdefined type: " + argType)
			} else {
				body += argName + " := map[" + arg.Type.Detail.String(false) + "]" + arg.Type.Container2Detail.String(false) + "{}\n"
				body += "copier.Copy(&" + argName + ", " + request_arg_name + "." + converted_arg_name + "}\n"
			}
		} else {
			argName := g.convertMessageFieldName(arg.Name)
			argNames = append(argNames, request_arg_name+"."+argName)
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
			argType, err := g.getGrpcTypeString(arg.Type)
			if err != nil {
				return "", nil
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				if v.IsEnum {
					body += "var " + new_retname + " " + g.appName + "." + v.Name + "\n"
					body += "copier.Copy(&" + new_retname + "," + retNames[idx] + ")\n"
				} else {
					body += new_retname + " := &" + g.appName + "." + v.Name + "{}" + "\n"
					body += "copier.Copy(" + new_retname + ", &" + retNames[idx] + ")\n"
				}
			}
			retNames[idx] = new_retname
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			new_retname := fmt.Sprintf("ret_updated%d", idx)
			argType, err := g.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + new_retname + " []*" + g.appName + "." + v.Name + "\n"
				body += "copier.Copy(&" + new_retname + ", &" + retNames[idx] + ")\n"
			}
			retNames[idx] = new_retname
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			new_retname := fmt.Sprintf("ret_updated%d", idx)
			argType, err := g.basicTypeToString(arg.Type.ContainerType2, arg.Type.Container2Detail)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += "var " + new_retname + "map[" + arg.Type.Detail.String(false) + "]*" + g.appName + "." + v.Name + "\n"
				body += "copier.Copy(&" + new_retname + ", &" + retNames[idx] + ")\n"
			}
			retNames[idx] = new_retname
		}
	}
	response_name := g.service_responses[service_name+":"+funcInfo.Name]
	if response_type, ok := g.responseTypes[response_name]; !ok {
		return "", errors.New("Failed to find response type for service " + service_name + " function " + funcInfo.Name)
	} else {
		new_retname := "response"
		body += new_retname + " := &" + g.appName + "." + response_type.Name + "{}\n"
		for idx, field := range response_type.Fields {
			body += new_retname + "." + field.Name + " = " + retNames[idx] + "\n"
		}
	}
	body += "return response," + retNames[len(retNames)-1]
	return body, nil
}

func (g *GRPCGenerator) GenerateServerMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, is_metrics_on bool, instance_name string) (map[string]string, error) {
	bodies := make(map[string]string)
	var methodInfos []MethodInfo
	var funcNames []string
	for name, method := range methods {
		funcNames = append(funcNames, name)
		response_name, err := g.packResponse(service_name, name, method.Return)
		if err != nil {
			return bodies, err
		}
		request_name, err := g.packRequest(service_name, name, method.Args)
		if err != nil {
			return bodies, err
		}
		request_arg_name := "request"
		body, err := g.generateServiceMethod(handler_name, service_name, method, request_arg_name, is_metrics_on)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
		var new_args []parser.ArgInfo
		var new_rets []parser.ArgInfo
		new_args = append(new_args, parser.GetContextArg("ctx"))
		method.Args = append(new_args, parser.GetPointerArg(request_arg_name, g.appName+"."+request_name))
		new_rets = append(new_rets, parser.GetPointerArg("", g.appName+"."+response_name))
		new_rets = append(new_rets, parser.GetErrorArg(""))
		method.Return = new_rets
		var method_string string
		method_string += "rpc " + name + " (" + request_name + ") returns (" + response_name + ") {}"
		methods[name] = method
		methodInfos = append(methodInfos, MethodInfo{Name: name, Val: method_string})
	}
	if is_metrics_on {
		// Add a metric method called startMetrics
		method, body := generateMetricMethod(handler_name, service_name, funcNames)
		methods[method.Name] = method
		bodies[method.Name] = body
	}
	run_method, run_body := g.generateServerRunMethod(instance_name, handler_name, service_name, is_metrics_on)
	methods[run_method.Name] = run_method
	bodies[run_method.Name] = run_body
	g.serviceTypes[service_name] = ServiceInfo{Name: service_name, Methods: methodInfos}
	g.functions[service_name] = funcNames
	return bodies, nil
}

func (g *GRPCGenerator) generateClientMethod(handler_name string, service_name string, funcInfo parser.FuncInfo, has_timeout bool) (string, error) {
	var body string
	var argNames []string
	request_name := g.service_requests[service_name+":"+funcInfo.Name]
	body += "request := &" + g.appName + "." + request_name + "{}\n"
	for idx, arg := range funcInfo.Args {
		argNames = append(argNames, arg.Name)
		if idx == 0 {
			// Ignore context argument if no timeout
			if has_timeout {
				body += arg.Name + ", cancel := context.WithTimeout(" + arg.Name + "," + handler_name + ".Timeout)\n"
				body += "defer cancel()\n"
			}
			continue
		}
		if arg.Type.BaseType == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := g.getGrpcTypeString(arg.Type)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				if v.IsEnum {
					body += "var " + argName + " " + g.appName + "." + v.Name + "\n"
					body += "copier.Copy(&" + argName + ", " + arg.Name + ")\n"
				} else {
					body += argName + " := &" + g.appName + "." + v.Name + "{}\n"
					body += "copier.Copy(" + argName + ", &" + arg.Name + ")\n"
					body += "request." + converted_arg_name + " = " + argName + "\n"
				}
			}
		} else if arg.Type.BaseType == parser.LIST && arg.Type.ContainerType1 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := g.basicTypeToString(arg.Type.ContainerType1, arg.Type.Detail)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += argName + " := []*" + g.appName + "." + v.Name + "{}\n"
				body += "copier.Copy(" + argName + ", &" + arg.Name + ")\n"
				body += "request." + converted_arg_name + " = " + argName + "\n"
			}
		} else if arg.Type.BaseType == parser.MAP && arg.Type.ContainerType2 == parser.USERDEFINED {
			argName := fmt.Sprintf("arg%d", idx)
			argNames = append(argNames, argName)
			argType, err := g.basicTypeToString(arg.Type.ContainerType2, arg.Type.Container2Detail)
			converted_arg_name := g.convertMessageFieldName(arg.Name)
			if err != nil {
				return "", err
			}
			if v, ok := g.remoteTypes[argType]; !ok {
				return "", errors.New("Unknown Userdefined type: " + argType)
			} else {
				body += argName + " := map[" + arg.Type.Detail.String(false) + "]*" + g.appName + "." + v.Name + "{}\n"
				body += "copier.Copy(" + argName + ", &" + arg.Name + ")\n"
				body += "request." + converted_arg_name + " = " + argName + "\n"
			}
		} else {
			argName := g.convertMessageFieldName(arg.Name)
			body += "request." + argName + " = " + arg.Name + "\n"
		}
	}
	if len(funcInfo.Return) != 1 {
		body += "response, err := " + handler_name + ".client." + funcInfo.Name + "(" + argNames[0] + ",request)\n"
	} else {
		body += "_, err := " + handler_name + ".client." + funcInfo.Name + "(" + argNames[0] + ",request)\n"
	}
	resp_name := g.service_responses[service_name+":"+funcInfo.Name]
	response_body := ""
	var retNames []string
	if respType, ok := g.responseTypes[resp_name]; !ok {
		return "", errors.New("Response object not found for function " + service_name + "." + funcInfo.Name)
	} else {
		for idx, field := range respType.Fields {
			if field.Type.BaseType == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := g.getGrpcTypeString(field.Type)
				if err != nil {
					return "", err
				}
				if _, ok := g.remoteTypes[argType]; !ok {
					log.Println("Unknowns Userdefined type in responsetype:", argType)
					return "", errors.New("Unknown Userdefined type: " + argType)
				} else {
					body += retName + " := " + field.Type.Detail.UserType + "{}\n"
					response_body += "copier.Copy(&" + retName + ", response" + "." + field.Name + ")\n"
				}
				retNames = append(retNames, retName)
			} else if field.Type.BaseType == parser.LIST && field.Type.ContainerType1 == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := g.basicTypeToString(field.Type.ContainerType1, field.Type.Detail)
				if err != nil {
					return "", err
				}
				if _, ok := g.remoteTypes[argType]; !ok {
					return "", errors.New("Unknown Userdefined type: " + argType)
				} else {
					body += retName + " := []" + field.Type.Detail.UserType + "{}\n"
					response_body += "copier.Copy(&" + retName + ", response." + field.Name + ")\n"
				}
				retNames = append(retNames, retName)
			} else if field.Type.BaseType == parser.MAP && field.Type.ContainerType2 == parser.USERDEFINED {
				retName := fmt.Sprintf("ret%d", idx)
				argType, err := g.basicTypeToString(field.Type.ContainerType2, field.Type.Container2Detail)
				if err != nil {
					return "", err
				}
				if _, ok := g.remoteTypes[argType]; !ok {
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

func (g *GRPCGenerator) GenerateClientMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, nextNodeMethodArgs []parser.ArgInfo, nextNodeMethodReturn []parser.ArgInfo, is_metrics_on bool, has_timeout bool) (map[string]string, error) {
	bodies := make(map[string]string)
	for name, method := range methods {
		method.Args = append(method.Args, nextNodeMethodArgs...)
		last_return := method.Return[len(method.Return)-1]
		method.Return = append(method.Return[:len(method.Return)-1], nextNodeMethodReturn...)
		method.Return = append(method.Return, last_return)
		methods[name] = method
		body, err := g.generateClientMethod(handler_name, service_name, method, has_timeout)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
	}
	return bodies, nil
}

func (g *GRPCGenerator) GenerateServerConstructor(prev_handler string, service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", handler_name)}
	args := []parser.ArgInfo{parser.GetPointerArg("old_handler", prev_handler)}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	fields := []parser.ArgInfo{parser.GetPointerArg("service", prev_handler), parser.GetBasicArg("", g.appName+".Unimplemented"+base_name+"Server")}
	var imports []parser.ImportInfo
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "os"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "net"})
	//imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "time"})
	//imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "google.golang.org/grpc/keepalive"})
	if is_metrics_on {
		funcs := g.functions[base_name]
		fields = append(fields, generateMetricFields(funcs)...)
		imports = append(imports, generateMetricImports()...)
	}
	body := ""
	body += "handler := &" + handler_name + "{service:old_handler}\n"
	body += "return handler"

	return funcInfo, body, imports, fields, []parser.StructInfo{}
}

func (g *GRPCGenerator) generateServerRunMethod(service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string) {
	var body string
	fn := parser.FuncInfo{Name: "Run", Args: []parser.ArgInfo{}, Return: []parser.ArgInfo{parser.GetErrorArg("")}}
	body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn errors.New(\"Address or Port were not set\")\n}\n"
	body += "lis, err := net.Listen(\"tcp\", addr + \":\" + port)\n"
	body += "if err != nil {\n\treturn err\n}\n"
	body += "grpcServer := grpc.NewServer()\n"
	body += g.appName + ".Register" + base_name + "Server(grpcServer," + handler_name + ")\n"
	if is_metrics_on {
		body += generateMetricConstructorBody(handler_name)
	}
	body += "return grpcServer.Serve(lis)\n"
	return fn, body
}

func (g *GRPCGenerator) GenerateClientConstructor(service_name string, handler_name string, base_name string, is_metrics_on bool, timeout string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", handler_name), parser.GetErrorArg("")}
	args := []parser.ArgInfo{}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	var imports []parser.ImportInfo
	fields := []parser.ArgInfo{parser.GetBasicArg("client", g.appName+"."+base_name+"Client")}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "os"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "google.golang.org/grpc/credentials/insecure"})

	resolver_name := ""
	has_resolver := false
	if v, ok := g.custom_params["resolver"]; ok {
		has_resolver = true
		resolver_name = v
	}
	body := ""
	if !has_resolver {
		body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
		body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	} else {
		body += "addr := os.Getenv(\"" + resolver_name + "_ADDRESS\")\n"
		body += "port := os.Getenv(\"" + resolver_name + "_PORT\")\n"
	}
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn nil, errors.New(\"Address or port were not set\")\n}\n"
	body += "var opts []grpc.DialOption\n"
	body += "opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))\n"
	retBody := "return &" + base_name + "RPCClient{client:client}, nil\n"
	if timeout != "" {
		body += "duration, err := time.ParseDuration(\"" + timeout + "\")\n"
		body += "if err == nil {\n"
		body += "\topts = append(opts, grpc.WithTimeout(duration))\n"
		body += "}\n"
		imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "time"})
		fields = append(fields, parser.GetBasicArg("Timeout", "time.Duration"))
		retBody = "return &" + base_name + "RPCClient{client:client, Timeout: duration}, nil\n"
	}
	conn_string := "addr + \":\" + port"
	if has_resolver {
		conn_string = "\"consul://\" + " + conn_string + " + \"/" + service_name + "\""
		imports = append(imports, parser.ImportInfo{ImportName: "_", FullName: "github.com/mbobakov/grpc-consul-resolver"})
		body += "opts = append(opts, grpc.WithDefaultServiceConfig(`{\"loadBalancingPolicy\": \"round_robin\"}`))\n"
	}
	body += "conn, err := grpc.Dial(" + conn_string + ", opts...)\n"
	body += "if err != nil {\n"
	body += "\treturn nil, err\n}\n"
	body += "client := " + g.appName + ".New" + base_name + "Client(conn)\n"
	body += retBody
	return funcInfo, body, imports, fields, []parser.StructInfo{}
}

func (g *GRPCGenerator) basicTypeToString(baseType parser.Type, typeDetail parser.TypeDetail) (string, error) {
	if baseType == parser.BASIC {
		if typeDetail.TypeName == parser.INT64 {
			return "int64", nil
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
	return "", errors.New("Unsupported type for grpc: " + baseType.String())
}

func (g *GRPCGenerator) getGrpcTypeString(typeInfo parser.TypeInfo) (string, error) {
	if typeInfo.BaseType == parser.BASIC || typeInfo.BaseType == parser.USERDEFINED {
		return g.basicTypeToString(typeInfo.BaseType, typeInfo.Detail)
	} else if typeInfo.BaseType == parser.MAP {
		basic_type1, err := g.basicTypeToString(typeInfo.ContainerType1, typeInfo.Detail)
		if err != nil {
			return "", err
		}
		basic_type2, err := g.basicTypeToString(typeInfo.ContainerType2, typeInfo.Container2Detail)
		if err != nil {
			return "", err
		}
		return "map<" + basic_type1 + "," + basic_type2 + ">", nil
	} else if typeInfo.BaseType == parser.LIST {
		basic_type, err := g.basicTypeToString(typeInfo.ContainerType1, typeInfo.Detail)
		if err != nil {
			return "", err
		}
		return "repeated " + basic_type, nil
	}
	return "", errors.New("Unsupported type for grpc: " + typeInfo.BaseType.String())
}

func (g *GRPCGenerator) SetCustomParameters(params map[string]string) {
	// Set custom parameters
	for k, v := range params {
		g.custom_params[k] = v
	}
}
