package netgen

import (
	"fmt"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type DefaultWebGenerator struct {
	appName           string
	remoteTypes       map[string]*parser.ImplInfo
	enums             map[string]*parser.EnumInfo
	functions         map[string][]string          // service_name -> list of functions
	response_objs     map[string]parser.StructInfo // response_name -> Detailed Info
	service_responses map[string]map[string]string // service_name -> func_name -> obj_name
}

func NewDefaultWebGenerator() NetworkGenerator {
	return &DefaultWebGenerator{functions: make(map[string][]string), response_objs: make(map[string]parser.StructInfo), service_responses: make(map[string]map[string]string)}
}

func (d *DefaultWebGenerator) GetRequirements() []parser.RequireInfo {
	return []parser.RequireInfo{parser.RequireInfo{Name: "github.com/gorilla/mux", Path: "", Version: "v1.8.0"}}
}

func (d *DefaultWebGenerator) packResponse(service_name string, func_name string, retVals []parser.ArgInfo) {
	if _, ok := d.service_responses[service_name]; !ok {
		d.service_responses[service_name] = make(map[string]string)
	}
	sr_map := d.service_responses[service_name]
	obj_name := service_name + "_" + func_name + "_WebResponse"
	sr_map[func_name] = obj_name
	var fields []parser.ArgInfo
	if len(retVals) == 1 {
		// No object needs to be created. Create an empty response.
		d.response_objs[obj_name] = parser.StructInfo{Name: obj_name, Fields: fields}
	} else {
		// (Ignore error field at the end)
		for idx := 0; idx < len(retVals)-1; idx += 1 {
			field := retVals[idx]
			field.Name = fmt.Sprintf("Ret%d", idx)
			fields = append(fields, field)
		}
		d.response_objs[obj_name] = parser.StructInfo{Name: obj_name, Fields: fields}
	}
}

func (d *DefaultWebGenerator) generateServiceMethod(handler_name string, service_name string, funcInfo parser.FuncInfo, is_metrics_on bool) (string, error) {
	var body string
	if is_metrics_on {
		body += generateFunctionWrapperBody(handler_name, funcInfo.Name)
	}
	if len(funcInfo.Args) > 1 {
		body += "var err error\n"
	}
	var arg_names []string
	for idx, arg := range funcInfo.Args {
		if idx == 0 {
			body += "ctx := context.Background()\n"
			arg_names = append(arg_names, "ctx")
			continue
		}
		body += "var " + arg.Name + " " + arg.Type.String() + "\n"
		arg_id := fmt.Sprintf("arg%d", idx)
		body += arg_id + " := r.FormValue(\"" + arg.Name + "\")\n"
		body += "if " + arg_id + " != \"\" {\n"
		body += "\terr = json.Unmarshal([]byte(" + arg_id + "), &" + arg.Name + ")\n"
		body += "\tif err != nil {\n"
		body += "\t\thttp.Error(w, err.Error(), 500)\n"
		body += "\t\tlog.Println(err)\n"
		body += "\t\tlog.Println(" + arg_id + `, "` + arg.Name + `")` + "\n"
		body += "\t\treturn\n"
		body += "\t}\n"
		body += "}\n"
		arg_names = append(arg_names, arg.Name)
	}
	var ret_names []string
	for idx, _ := range funcInfo.Return {
		ret_name := fmt.Sprintf("ret%d", idx)
		ret_names = append(ret_names, ret_name)
	}
	body += strings.Join(ret_names, ", ") + " := " + handler_name + ".service." + funcInfo.Name + "(" + strings.Join(arg_names, ", ") + ")\n"
	// Return an error if the internal service returns an error
	last_ret_name := ret_names[len(ret_names)-1]
	body += "if " + last_ret_name + " != nil {\n"
	body += "\thttp.Error(w, " + last_ret_name + ".Error(), 500)\n\treturn\n}\n"
	// Result should be encoded with json in the response writer
	// Get correct response name
	service_map := d.service_responses[service_name]
	response_name := service_map[funcInfo.Name]
	body += "response := " + response_name + "{}\n"
	for idx, ret_name := range ret_names {
		response_ret_name := fmt.Sprintf("Ret%d", idx)
		if idx == len(ret_names)-1 {
			// Last return is an error so we don't need to encode that response
			break
		}
		body += "response." + response_ret_name + " = " + ret_name + "\n"
	}
	body += "json.NewEncoder(w).Encode(response)\n"

	return body, nil
}

func (d *DefaultWebGenerator) GenerateServerMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, is_metrics_on bool, instance_name string) (map[string]string, error) {
	bodies := make(map[string]string)
	var func_names []string
	for name, method := range methods {
		// Create Response Object
		d.packResponse(service_name, name, method.Return)
		func_names = append(func_names, name)
		body, err := d.generateServiceMethod(handler_name, service_name, method, is_metrics_on)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
		var new_args []parser.ArgInfo
		new_args = append(new_args, parser.GetBasicArg("w", "http.ResponseWriter"))
		new_args = append(new_args, parser.GetPointerArg("r", "http.Request"))
		method.Args = new_args
		method.Return = []parser.ArgInfo{}
		methods[name] = method
	}
	if is_metrics_on {
		// Add a metric method called startMetrics
		method, body := generateMetricMethod(handler_name, service_name, func_names)
		methods[method.Name] = method
		bodies[method.Name] = body
	}
	d.functions[service_name] = func_names
	run_method, run_body := d.generateServerRunMethod(instance_name, handler_name, service_name, is_metrics_on)
	methods[run_method.Name] = run_method
	bodies[run_method.Name] = run_body
	return bodies, nil
}

func (d *DefaultWebGenerator) generateClientMethod(handler_name string, service_name string, funcInfo parser.FuncInfo) (string, error) {
	var body string
	body += "values := url.Values{}\n"
	for idx, arg := range funcInfo.Args {
		if idx == 0 {
			// Ignore context arg
			continue
		}
		argname := fmt.Sprintf("arg%d", idx)
		body += argname + ", _ := json.Marshal(" + arg.Name + ")\n"
		body += "values.Set(" + arg.Name + ", string(" + argname + "))\n"
	}
	body += "resp, err := http.PostForm(" + handler_name + ".url" + ", values)\n"
	var ret_names []string
	var err_ret_names []string
	for idx, retarg := range funcInfo.Return {
		if idx == len(funcInfo.Return)-1 {
			continue
		}
		ret_names = append(ret_names, fmt.Sprintf("response.Ret%d", idx))
		err_ret_names = append(err_ret_names, fmt.Sprintf("var ret%d %s", idx, retarg.Type.String()))
	}
	// Check error
	body += "if err != nil {\n"
	body += "\treturn " + strings.Join(err_ret_names, ", ") + ", err\n"
	body += "}\n"
	body += "defer resp.Body.Close()\n"
	// Get correct response name
	service_map := d.service_responses[service_name]
	response_name := service_map[funcInfo.Name]
	body += "var response " + response_name
	body += "json.NewDecoder(resp.Body).Decode(&response)\n"
	body += "return " + strings.Join(ret_names, ", ") + ", nil\n"
	return body, nil
}

func (d *DefaultWebGenerator) GenerateClientMethods(handler_name string, service_name string, methods map[string]parser.FuncInfo, nextNodeMethodArgs []parser.ArgInfo, nextNodeMethodReturn []parser.ArgInfo, is_metrics_on bool, has_timeout bool) (map[string]string, error) {
	bodies := make(map[string]string)
	for name, method := range methods {
		method.Args = append(method.Args, nextNodeMethodArgs...)
		last_return := method.Return[len(method.Return)-1]
		method.Return = append(method.Return[:len(method.Return)-1], nextNodeMethodReturn...)
		method.Return = append(method.Return, last_return)
		methods[name] = method
		body, err := d.generateClientMethod(handler_name, service_name, method)
		if err != nil {
			return bodies, err
		}
		bodies[name] = body
	}
	return bodies, nil
}

func (d *DefaultWebGenerator) SetAppName(appName string) {
	if d.appName == "" {
		d.appName = appName
	}
}

func (d *DefaultWebGenerator) GenerateFiles(outdir string) error {
	// Web generators have no extra files
	return nil
}

func (d *DefaultWebGenerator) ConvertRemoteTypes(remoteTypes map[string]*parser.ImplInfo) error {
	d.remoteTypes = remoteTypes
	return nil
}

func (d *DefaultWebGenerator) ConvertEnumTypes(enumTypes map[string]*parser.EnumInfo) error {
	d.enums = enumTypes
	return nil
}

func (d *DefaultWebGenerator) GetImports(_ bool) []parser.ImportInfo {
	return []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: "github.com/gorilla/mux"}, parser.ImportInfo{ImportName: "", FullName: "net/http"}, parser.ImportInfo{ImportName: "", FullName: "os"}, parser.ImportInfo{ImportName: "", FullName: "context"}, parser.ImportInfo{ImportName: "", FullName: "encoding/json"}}
}

func (d *DefaultWebGenerator) getClientImports() []parser.ImportInfo {
	imports := d.GetImports(false)
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "net/url"})
	return imports
}

func (d *DefaultWebGenerator) GenerateServerConstructor(prev_handler string, service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", handler_name)}
	args := []parser.ArgInfo{parser.GetPointerArg("old_handler", prev_handler)}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	fields := []parser.ArgInfo{parser.GetPointerArg("service", prev_handler)}
	fields = append(fields, parser.GetBasicArg("url", "string"))
	imports := []parser.ImportInfo{}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "log"})
	body := ""
	body += "handler := &" + handler_name + "{service: old_handler, url: url}\n"
	var structs []parser.StructInfo
	for _, v := range d.service_responses[base_name] {
		structs = append(structs, d.response_objs[v])
	}
	if is_metrics_on {
		funcs := d.functions[base_name]
		fields = append(fields, generateMetricFields(funcs)...)
		imports = append(imports, generateMetricImports()...)
	}
	body = "handler := &" + handler_name + "{service: old_handler, url: \"\"}\n"
	body += "return handler"
	return funcInfo, body, imports, fields, structs
}

func (d *DefaultWebGenerator) generateServerRunMethod(service_name string, handler_name string, base_name string, is_metrics_on bool) (parser.FuncInfo, string) {
	var body string
	fn := parser.FuncInfo{Name: "Run", Args: []parser.ArgInfo{}, Return: []parser.ArgInfo{parser.GetErrorArg("")}}

	body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn errors.New(\"Address or Port were not set\")\n}\n"
	body += "url := \"http://\" + addr + \":\" + port\n"
	body += "router := mux.NewRouter()\n"

	body += handler_name + ".url = url\n"
	funcs := d.functions[base_name]
	for _, f := range funcs {
		body += "router.Path(\"/" + f + "\").HandlerFunc(" + handler_name + "." + f + ")\n"
	}

	if is_metrics_on {
		body += generateMetricConstructorBody(handler_name)
	}
	body += "log.Println(\"Launching Server\")\n"
	body += "return http.ListenAndServe(addr + \":\" + port, router)"
	return fn, body
}

func (d *DefaultWebGenerator) GenerateClientConstructor(service_name string, handler_name string, base_name string, is_metrics_on bool, timeout string) (parser.FuncInfo, string, []parser.ImportInfo, []parser.ArgInfo, []parser.StructInfo) {
	func_name := "New" + handler_name
	ret_args := []parser.ArgInfo{parser.GetPointerArg("", handler_name), parser.GetErrorArg("")}
	args := []parser.ArgInfo{}
	funcInfo := parser.FuncInfo{Name: func_name, Args: args, Return: ret_args}
	imports := d.getClientImports()
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "errors"})
	fields := []parser.ArgInfo{parser.GetBasicArg("url", "string")}
	body := ""
	body += "addr := os.Getenv(\"" + service_name + "_ADDRESS\")\n"
	body += "port := os.Getenv(\"" + service_name + "_PORT\")\n"
	body += "if addr == \"\" || port == \"\" {\n"
	body += "\treturn nil, errors.New(\"Address or port were not set\")\n}\n"
	body += "url := \"http://\" + addr + \":\" + port\n"
	body += "return &" + base_name + "WebClient{url: url}, nil\n"
	var structs []parser.StructInfo
	for _, v := range d.service_responses[base_name] {
		structs = append(structs, d.response_objs[v])
	}
	return funcInfo, body, imports, fields, structs
}
