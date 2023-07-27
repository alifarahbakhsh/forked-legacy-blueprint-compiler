package generators

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/netgen"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"

	"golang.org/x/mod/modfile"
)

type MainVisitor struct {
	DefaultVisitor
	logger        *log.Logger
	out_dir       string
	curDir        string
	pkgName       string
	runNames      []string
	ctrName       string
	pathpkgs      map[string]string
	impls         map[string]*parser.ImplInfo
	specDir       string
	depgenfactory *deploy.DeployerGeneratorFactory
	cur_env_vars  map[string]string
	public_ports  map[int]int
	address       string
	hostname      string
	port          int
	imageName     string
	isservice     bool
	addrs         map[string]ConnInfo
	inventory     []parser.Node
	deployInfo    *deploy.DeployInfo
	// Client Constructor Generation state
	nextClientNode   *ServiceImplInfo
	curClientNode    *ServiceImplInfo
	curBody          string
	prev_client_name string
	client_names     map[string]string
	client_imports   []parser.ImportInfo
	added_imports    map[string]bool
	frameworks       map[string]netgen.NetworkGenerator
	DepGraph         *DependencyGraph
	// Process Main Function state
	localServicesInfo map[string]map[string]string
	ProcInfo          *ProcessRunServicesInfo
	curProcName       string
	localServices     map[string]string
}

type ProcessRunServicesInfo struct {
	GetFuncNames  map[string]string
	GetFuncArgs   map[string][]parser.ArgInfo
	InstanceTypes map[string]parser.ArgInfo
	RunFuncArgs   map[string][]parser.ArgInfo
	Order         []string
}

func NewProcessRunServicesInfo() *ProcessRunServicesInfo {
	return &ProcessRunServicesInfo{GetFuncNames: make(map[string]string), GetFuncArgs: make(map[string][]parser.ArgInfo), InstanceTypes: make(map[string]parser.ArgInfo), RunFuncArgs: make(map[string][]parser.ArgInfo)}
}

func NewMainVisitor(logger *log.Logger, out_dir string, pathpkgs map[string]string, impls map[string]*parser.ImplInfo, specDir string, depgenfactory *deploy.DeployerGeneratorFactory, addrs map[string]ConnInfo, inventory []parser.Node, frameworks map[string]netgen.NetworkGenerator, dg *DependencyGraph, localServicesInfo map[string]map[string]string) *MainVisitor {
	return &MainVisitor{logger: logger, out_dir: out_dir, curDir: out_dir, pathpkgs: pathpkgs, impls: impls, specDir: specDir, depgenfactory: depgenfactory, addrs: addrs, inventory: inventory, frameworks: frameworks, DepGraph: dg, localServicesInfo: localServicesInfo}
}

func (v *MainVisitor) modifySpecModFile() {
	outspec_dir := path.Join(v.out_dir, "spec")
	out_mod_file := path.Join(outspec_dir, "go.mod")
	v.logger.Println("Reading mod file:", out_mod_file)
	data, err := ioutil.ReadFile(out_mod_file)
	if err != nil {
		v.logger.Fatal(err)
	}
	f, err := modfile.ParseLax(out_mod_file, data, nil)
	if err != nil {
		v.logger.Fatal(err)
	}
	f.Module.Mod.Path = "spec"
	f.Module.Syntax.Token = []string{"module", "spec"}
	err = f.AddRequire("gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler", "v0.0.1")
	if err != nil {
		v.logger.Fatal(err)
	}
	bytes, err := f.Format()
	if err != nil {
		v.logger.Fatal(err)
	}
	outf, err := os.OpenFile(out_mod_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	defer outf.Close()
	_, err = outf.Write(bytes)
	if err != nil {
		v.logger.Fatal(err)
	}
	prev_dir, err := os.Getwd()
	if err != nil {
		v.logger.Fatal(err)
	}
	err = os.Chdir(outspec_dir)
	if err != nil {
		v.logger.Fatal(err)
	}
	cmd := exec.Command("go", "mod", "tidy")
	out, err := cmd.CombinedOutput()
	if err != nil {
		v.logger.Fatal(string(out))
	}
	v.logger.Println(string(out))
	err = os.Chdir(prev_dir)
	if err != nil {
		v.logger.Fatal(err)
	}
}

func (v *MainVisitor) copyEnvVars(envVars map[string]string) {
	for key, val := range envVars {
		v.cur_env_vars[key] = val
	}
}

func (v *MainVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.modifySpecModFile()
	v.logger.Println("Starting MainVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending MainVisitor visit")
	// Generate docker-compose file + other config files (Kubernetes, OpenShift)

	for _, depgen := range v.depgenfactory.Generators {
		fp, err := filepath.Abs(v.out_dir)
		if err != nil {
			v.logger.Fatal(err)
		}
		err = depgen.GenerateConfigFiles(fp)
		if err != nil {
			v.logger.Fatal(err)
		}
	}
}

func (v *MainVisitor) VisitAnsibleContainerNode(_ Visitor, n *AnsibleContainerNode) {

	v.DefaultVisitor.VisitAnsibleContainerNode(v, n)
	depgen, err := v.depgenfactory.GetGenerator("ansible")

	depgen.(*deploy.AnsibleDeployerGenerator).SetInventory(v.inventory)
	if err != nil {
		v.logger.Fatal(err)
	}

	v.deployInfo.Hostname = v.hostname
	if !v.isservice {
		depgen.AddChoice(n.Name, v.deployInfo)
	} else {
		depgen.AddService(n.Name, v.deployInfo)
	}
}

func (v *MainVisitor) VisitKubernetesContainerNode(_ Visitor, n *KubernetesContainerNode) {
	v.DefaultVisitor.VisitKubernetesContainerNode(v, n)
	// Ensure that the generator is initialized for config file generation
	_, err := v.depgenfactory.GetGenerator("kubernetes")
	if err != nil {
		v.logger.Fatal(err)
	}
}

func (v *MainVisitor) VisitDockerContainerNode(_ Visitor, n *DockerContainerNode) {
	oldPath := v.curDir
	v.curDir = path.Join(oldPath, strings.ToLower(n.Name))

	v.ctrName = strings.ToLower(n.Name)
	v.cur_env_vars = make(map[string]string)
	v.public_ports = make(map[int]int)
	v.isservice = true
	v.imageName = ""
	v.DefaultVisitor.VisitDockerContainerNode(v, n)
	// Generate Docker File for each container

	docker_dir := path.Join(v.curDir, "docker")
	err := os.MkdirAll(docker_dir, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	v.generateModFile(v.curDir, n)
	v.generateDockerFile(docker_dir, n)

	if !v.isservice {
		dockerInfo := &deploy.DeployInfo{Address: v.address, Port: v.port, DockerPath: "", ImageName: v.imageName, EnvVars: v.cur_env_vars, PublicPorts: v.public_ports}
		v.deployInfo = dockerInfo
		depgen, err := v.depgenfactory.GetGenerator("docker")
		if err != nil {
			v.logger.Fatal(err)
		}
		depgen.AddChoice(n.Name, dockerInfo)
	}
	v.curDir = oldPath
}

func (v *MainVisitor) generateModFile(ctr_dir string, n *DockerContainerNode) {
	service_nodes := n.GetNodes("FuncServiceNode")
	queue_service_nodes := n.GetNodes("QueueServiceNode")
	if len(service_nodes) == 0 && len(queue_service_nodes) == 0 {
		return
	}
	mod_name := strings.ToLower(n.Name)
	filename := path.Join(ctr_dir, "go.mod")
	data := []byte("module " + mod_name + "\n\ngo 1.18\n\n")
	f, err := modfile.ParseLax(filename, data, nil)
	if err != nil {
		v.logger.Fatal(err)
	}
	err = f.AddRequire("gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler", "v0.0.1")
	if err != nil {
		v.logger.Fatal(err)
	}
	err = f.AddRequire("spec", "v1.0.0")
	if err != nil {
		v.logger.Fatal(err)
	}
	// Add Require for generated folders (THRIFT)
	var requires []parser.RequireInfo
	for name, netgenerator := range v.frameworks {
		v.logger.Println("Getting requirementes for ", name)
		framework_requires := netgenerator.GetRequirements()
		requires = append(requires, framework_requires...)
	}
	for _, require := range requires {
		v.logger.Println("Adding require ", require.Name)
		err = f.AddRequire(require.Name, require.Version)
		if err != nil {
			v.logger.Fatal(err)
		}
	}
	for _, replace := range requires {
		if replace.Name == replace.Path {
			continue
		}
		err = f.AddReplace(replace.Name, replace.Version, replace.Path, "")
		if err != nil {
			v.logger.Fatal(err)
		}
	}
	err = f.AddReplace("spec", "", "../spec", "")
	if err != nil {
		v.logger.Fatal(err)
	}
	bytes, err := f.Format()
	if err != nil {
		v.logger.Fatal(err)
	}
	outf, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	defer outf.Close()
	_, err = outf.Write(bytes)
	if err != nil {
		v.logger.Fatal(err)
	}
	prev_dir, err := os.Getwd()
	if err != nil {
		v.logger.Fatal(err)
	}
	err = os.Chdir(ctr_dir)
	if err != nil {
		v.logger.Fatal(err)
	}
	err = os.Chdir(prev_dir)
	if err != nil {
		v.logger.Fatal(err)
	}
}

func (v *MainVisitor) generateDockerFile(docker_dir string, n *DockerContainerNode) {
	service_nodes := n.GetNodes("FuncServiceNode")
	queue_service_nodes := n.GetNodes("QueueServiceNode")
	if len(service_nodes) == 0 && len(queue_service_nodes) == 0 {
		return
	}
	dockerfile := path.Join(docker_dir, "Dockerfile")
	outf, err := os.OpenFile(dockerfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	name := strings.ToLower(n.Name)
	docker_string := ""
	docker_string += "FROM golang:1.18-buster AS build\n\n"
	docker_string += "WORKDIR /app\n\n"
	docker_string += "COPY ./ ./\n\n"
	docker_string += "WORKDIR /app/spec\n"
	docker_string += "RUN go mod download\n\n"
	docker_string += "WORKDIR /app/" + name + "\n"
	docker_string += "RUN go mod download\n\n"
	docker_string += "WORKDIR /app/" + name + "/app\n"
	docker_string += "RUN go mod tidy\n"
	docker_string += "RUN go build -o /" + name + "\n"
	docker_string += "FROM gcr.io/distroless/base-debian10\n"
	docker_string += "WORKDIR /\n"
	docker_string += "COPY --from=build " + name + " " + name + "\n"
	docker_string += "ENTRYPOINT [\"/" + name + "\"]\n\n"

	_, err = outf.WriteString(docker_string)
	if err != nil {
		v.logger.Fatal(err)
	}

	dockerInfo := &deploy.DeployInfo{Address: v.address, Port: v.port, DockerPath: path.Join(name, "docker"), ImageName: v.imageName, EnvVars: v.cur_env_vars, PublicPorts: v.public_ports, NumReplicas: v.deployInfo.NumReplicas}
	v.deployInfo = dockerInfo
	depgen, err := v.depgenfactory.GetGenerator("docker")
	if err != nil {
		v.logger.Fatal(err)
	}
	depgen.AddService(n.Name, dockerInfo)
}

func (v *MainVisitor) generateMainFile() {
	out_dir := path.Join(v.curDir, "app")
	err := os.MkdirAll(out_dir, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	out_file := path.Join(out_dir, "main.go")
	outf, err := os.OpenFile(out_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	defer outf.Close()
	_, err = outf.WriteString("// Blueprint: auto-generated by Blueprint core\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	_, err = outf.WriteString("package main\n\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	_, err = outf.WriteString("import \"" + v.ctrName + "/" + v.pkgName + "\"\nimport \"sync\"\nimport \"log\"\n\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	func_body := "func main() {\n"
	for name, arg := range v.ProcInfo.InstanceTypes {
		func_body += "\tvar " + name + " *" + v.pkgName + "." + arg.String() + "\n"
	}
	for _, name := range v.ProcInfo.Order {
		var arg_strings []string
		fn_name := v.ProcInfo.GetFuncNames[name]
		fn_args := v.ProcInfo.GetFuncArgs[name]
		for _, arg := range fn_args {
			arg_strings = append(arg_strings, arg.Name)
		}
		func_body += "\t" + name + " = " + v.pkgName + "." + fn_name + "(" + strings.Join(arg_strings, ", ") + ")\n"
	}
	func_body += "\tc := make(chan error, 1)\n"
	func_body += "\twg_done := make(chan bool)\n"
	func_body += "\tvar wg sync.WaitGroup\n"
	func_body += "\twg.Add(" + strconv.Itoa(len(v.runNames)) + ")\n"
	for _, name := range v.runNames {
		func_body += "\tgo func(){\n"
		func_body += "\t\tdefer wg.Done()\n"
		ret_args := v.ProcInfo.RunFuncArgs[name]
		var ret_arg_strings []string
		for _, arg := range ret_args {
			ret_arg_strings = append(ret_arg_strings, arg.String())
		}
		func_body += "\t\terr := " + v.pkgName + "." + name + "(" + strings.Join(ret_arg_strings, ", ") + ")\n"
		func_body += "\t\tif err != nil {\n"
		func_body += "\t\t\tc <- err\n"
		func_body += "\t\t}\n"
		func_body += "\t}()\n"
	}
	func_body += "\tgo func(){\n"
	func_body += "\t\twg.Wait()\n"
	func_body += "\t\twg_done <- true\n"
	func_body += "\t}()\n"
	func_body += "\tselect {\n"
	func_body += "\tcase err := <- c:\n"
	func_body += "\t\tlog.Fatal(err)\n"
	func_body += "\tcase <- wg_done:\n"
	func_body += "\t\tlog.Println(\"Success\")\n"
	func_body += "\t}\n\n"
	func_body += "}\n"

	_, err = outf.WriteString(func_body)
	if err != nil {
		v.logger.Fatal(err)
	}
}

func (v *MainVisitor) VisitProcessNode(_ Visitor, n *ProcessNode) {
	oldPath := v.curDir
	v.curDir = path.Join(oldPath, strings.ToLower(n.Name))
	v.pkgName = strings.ToLower(n.Name)
	v.runNames = []string{}
	v.curProcName = n.Name
	v.localServices = make(map[string]string)
	if names, ok := v.localServicesInfo[n.Name]; ok {
		for name, val := range names {
			v.localServices[name] = val
		}
	}
	v.ProcInfo = NewProcessRunServicesInfo()
	v.DefaultVisitor.VisitProcessNode(v, n)
	// Generate a run_script for each process
	v.curDir = oldPath
	v.generateMainFile()
	v.pkgName = ""
}

func (v *MainVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.logger.Println("Visitng function node for", n.Name)
	v.address = n.DepInfo.Address
	v.port = n.DepInfo.Port
	v.hostname = n.DepInfo.Hostname
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	// Generate a function that starts the server for this service!
	out_file := path.Join(v.curDir, n.Name+".go")
	outf, err := os.OpenFile(out_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	defer outf.Close()
	_, err = outf.WriteString("// Blueprint: auto-generated by Blueprint Core\n")
	if err != nil {
		v.logger.Fatal()
	}
	_, err = outf.WriteString("package " + v.pkgName + "\n\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	func_name := "Get" + n.Name
	args := []parser.ArgInfo{}
	body := ""

	v.curClientNode = nil
	v.nextClientNode = nil
	v.curBody = ""
	v.prev_client_name = ""
	v.client_imports = []parser.ImportInfo{}
	v.added_imports = make(map[string]bool)
	// Maps the instanceName to its corresponding variable name
	client_names := make(map[string]string)
	v.copyEnvVars(n.DepInfo.EnvVars)
	for name, client_nodes := range n.ModifierClientNodes {
		conn_info := v.addrs[name]
		v.cur_env_vars[name+"_ADDRESS"] = conn_info.Address
		v.cur_env_vars[name+"_PORT"] = strconv.Itoa(conn_info.Port)
		prev_client_name := ""
		for i := len(client_nodes) - 1; i >= 0; i -= 1 {
			cnode := client_nodes[i]
			client_name := strings.ToLower(cnode.Name)
			// TODO: Use constructor params
			for _, arg := range cnode.Constructors[0].Args {
				v.logger.Println(arg.Name)
			}
			client_names[cnode.Name] = client_name
			body += client_name + " := " + cnode.Constructors[0].Name + "(" + prev_client_name + ")\n"
			prev_client_name = client_name
		}
	}

	v.client_names = client_names

	topo_ordering := v.DepGraph.Order[n.Name]
	for _, name := range topo_ordering {
		if val, ok := v.localServices[name]; ok {
			v.client_names[name] = name
			args = append(args, parser.GetPointerArg(name, val))
		}
		client_nodes := n.ParamClientNodes[name]
		v.logger.Println("Getting client nodes for", name)
		conn_info := v.addrs[name]
		v.logger.Println("Client nodes for", name, len(client_nodes))
		v.cur_env_vars[name+"_ADDRESS"] = conn_info.Address
		v.cur_env_vars[name+"_PORT"] = strconv.Itoa(conn_info.Port)
		for i := len(client_nodes) - 1; i >= 1; i -= 1 {
			cnode := client_nodes[i]
			v.curClientNode = cnode
			if cnode.ModifierNode != nil {
				cnode.ModifierNode.Accept(v)
			} else {
				client_name := strings.ToLower(cnode.Name)
				v.prev_client_name = client_name
			}
			v.nextClientNode = cnode
		}
		// Add default client constructor
		if len(client_nodes) != 1 {
			def_client_node := client_nodes[0]
			prev_client_node := client_nodes[1]
			cons_args := []parser.ArgInfo{parser.GetPointerArg("client", prev_client_node.Name)}
			cons_ret_args := []parser.ArgInfo{parser.GetPointerArg("", def_client_node.Name)}
			constructor := parser.FuncInfo{Name: "New" + def_client_node.Name, Args: cons_args, Return: cons_ret_args}
			cons_body := ""
			cons_body = "return &" + def_client_node.Name + "{client:client}"
			def_client_node.MethodBodies[constructor.Name] = cons_body
			def_client_node.Constructors = []parser.FuncInfo{constructor}
			def_client_node.Fields = []parser.ArgInfo{parser.GetPointerArg("client", prev_client_node.Name)}
			arg_strings := []string{v.prev_client_name}
			client_name := v.getVariableName(def_client_node)
			for idx, value := range def_client_node.Values {
				if idx == 0 {
					continue
				}
				if val, ok := v.client_names[value]; !ok {
					arg_strings = append(arg_strings, "\""+value+"\"")
				} else {
					arg_strings = append(arg_strings, val)
				}
			}
			v.curBody += client_name + " := " + def_client_node.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
			v.prev_client_name = client_name
		} else {
			// There is no previous client name
			def_client_node := client_nodes[0]
			arg_strings := []string{}
			client_name := v.getVariableName(def_client_node)
			for _, value := range def_client_node.Values {
				if val, ok := v.client_names[value]; !ok {
					arg_strings = append(arg_strings, "\""+value+"\"")
				} else {
					arg_strings = append(arg_strings, val)
				}
			}
			v.curBody += client_name + " := " + def_client_node.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
			v.prev_client_name = client_name
		}
		body += v.curBody
		v.curBody = ""
		client_names[name] = v.prev_client_name
	}

	handler_node := n.ASTServerNodes[0]
	var harg_strings []string
	for _, param := range n.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			harg_strings = append(harg_strings, "\""+ptype.Value+"\"")
		case *InstanceParameter:
			harg_strings = append(harg_strings, client_names[ptype.Name])
		}
	}

	implType := v.impls[handler_node.Name]
	pkg_name := v.pathpkgs[implType.PkgPath]
	import_path := "spec" + strings.ReplaceAll(implType.PkgPath, v.specDir, "")
	imports := []parser.ImportInfo{}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: import_path})
	imports = append(imports, v.client_imports...)

	var import_string string
	for _, imp := range imports {
		import_string += "import \"" + imp.FullName + "\"\n"
	}
	_, err = outf.WriteString(import_string)
	if err != nil {
		v.logger.Fatal(err)
	}

	main_handler_string := "spec_handler := " + pkg_name + "." + handler_node.Constructors[0].Name + "(" + strings.Join(harg_strings, ", ") + ")\n"
	body += main_handler_string
	prev_handler := "spec_handler"
	last_type := ""
	base_type := ""
	has_run_func := false
	for i := 0; i < len(n.ASTServerNodes); i += 1 {
		handler_node := n.ASTServerNodes[i]
		handler_name := strings.ToLower(handler_node.Name)
		// Use constructor params
		arg_strings := []string{prev_handler}
		for _, v := range handler_node.Values {
			if val, ok := client_names[v]; !ok {
				arg_strings = append(arg_strings, "\""+v+"\"")
			} else {
				arg_strings = append(arg_strings, val)
			}
		}
		body += handler_name + " := " + handler_node.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
		prev_handler = handler_name
		last_type = handler_node.Name
		base_type = handler_node.Name
		if _, ok := handler_node.Methods["Run"]; ok {
			has_run_func = ok
		}
	}
	body += "return " + prev_handler

	ret_args := []parser.ArgInfo{parser.GetPointerArg("", last_type)}

	var arg_strings []string
	for _, arg := range args {
		arg_strings = append(arg_strings, arg.String())
	}
	var ret_strings []string
	for _, arg := range ret_args {
		ret_strings = append(ret_strings, arg.String())
	}
	func_string := "func " + func_name + "(" + strings.Join(arg_strings, ", ") + ") "
	if len(ret_strings) > 1 {
		func_string += "(" + strings.Join(ret_strings, ", ") + ")"
	} else {
		func_string += strings.Join(ret_strings, ", ")
	}
	v.ProcInfo.GetFuncNames[n.Name] = func_name
	v.ProcInfo.Order = append(v.ProcInfo.Order, n.Name)
	v.ProcInfo.InstanceTypes[n.Name] = parser.GetBasicArg("", base_type)
	v.ProcInfo.GetFuncArgs[n.Name] = args
	func_string += " {\n"
	func_string += "\t" + strings.ReplaceAll(body, "\n", "\n\t")
	func_string += "\n}\n"
	_, err = outf.WriteString(func_string + "\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	if has_run_func {
		func_name := "Run" + n.Name
		args := []parser.ArgInfo{parser.GetPointerArg("service", last_type)}
		ret_args := []parser.ArgInfo{parser.GetErrorArg("")}
		var arg_strings []string
		for _, arg := range args {
			arg_strings = append(arg_strings, arg.String())
		}
		var ret_strings []string
		for _, arg := range ret_args {
			ret_strings = append(ret_strings, arg.String())
		}
		func_string := "func " + func_name + "(" + strings.Join(arg_strings, ", ") + ") " + strings.Join(ret_strings, ", ") + "{\n"
		func_string += "return service.Run()\n"
		func_string += "}\n"
		_, err = outf.WriteString(func_string + "\n")
		if err != nil {
			v.logger.Fatal(err)
		}
		n.RunFuncName = func_name
		v.runNames = append(v.runNames, func_name)
		v.ProcInfo.RunFuncArgs[func_name] = []parser.ArgInfo{parser.GetBasicArg(n.Name, "")}
	}
	v.deployInfo = n.DepInfo
}

func (v *MainVisitor) VisitQueueServiceNode(_ Visitor, n *QueueServiceNode) {
	v.logger.Println("Visiting queueservice node for", n.Name)
	v.address = n.DepInfo.Address
	v.port = n.DepInfo.Port
	v.hostname = n.DepInfo.Hostname
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	// Generate a function that starts the server for this service!
	out_file := path.Join(v.curDir, n.Name+".go")
	outf, err := os.OpenFile(out_file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		v.logger.Fatal(err)
	}
	defer outf.Close()
	_, err = outf.WriteString("package " + v.pkgName + "\n\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	func_name := "Run" + n.Name
	args := []parser.ArgInfo{}
	ret_args := []parser.ArgInfo{parser.GetErrorArg("")}
	body := ""
	v.curClientNode = nil
	v.nextClientNode = nil
	v.curBody = ""
	v.prev_client_name = ""
	v.client_imports = []parser.ImportInfo{}
	v.added_imports = make(map[string]bool)
	client_names := make(map[string]string)

	v.copyEnvVars(n.DepInfo.EnvVars)
	for name, client_nodes := range n.ModifierClientNodes {
		conn_info := v.addrs[name]
		v.cur_env_vars[name+"_ADDRESS"] = conn_info.Address
		v.cur_env_vars[name+"_PORT"] = strconv.Itoa(conn_info.Port)
		prev_client_name := ""
		for i := len(client_nodes) - 1; i >= 0; i -= 1 {
			cnode := client_nodes[i]
			client_name := strings.ToLower(cnode.Name)
			// TODO: Use constructor params
			for _, arg := range cnode.Constructors[0].Args {
				v.logger.Println(arg.Name)
			}
			client_names[cnode.Name] = client_name
			body += client_name + " := " + cnode.Constructors[0].Name + "(" + prev_client_name + ")\n"
			prev_client_name = client_name
		}
	}

	v.client_names = client_names

	topo_ordering := v.DepGraph.Order[n.Name]
	for _, name := range topo_ordering {
		client_nodes := n.ParamClientNodes[name]
		v.logger.Println("Getting client nodes for", name)
		conn_info := v.addrs[name]
		v.logger.Println("Client nodes for", name)
		v.cur_env_vars[name+"_ADDRESS"] = conn_info.Address
		v.cur_env_vars[name+"_PORT"] = strconv.Itoa(conn_info.Port)
		for i := len(client_nodes) - 1; i >= 1; i -= 1 {
			cnode := client_nodes[i]
			v.curClientNode = cnode
			if cnode.ModifierNode != nil {
				cnode.ModifierNode.Accept(v)
			} else {
				client_name := strings.ToLower(cnode.Name)
				v.prev_client_name = client_name
			}
			v.nextClientNode = cnode
		}
		// Add default client constructor
		if len(client_nodes) != 1 {
			def_client_node := client_nodes[0]
			prev_client_node := client_nodes[1]
			cons_args := []parser.ArgInfo{parser.GetPointerArg("client", prev_client_node.Name)}
			cons_ret_args := []parser.ArgInfo{parser.GetPointerArg("", def_client_node.Name)}
			constructor := parser.FuncInfo{Name: "New" + def_client_node.Name, Args: cons_args, Return: cons_ret_args}
			cons_body := ""
			cons_body = "return &" + def_client_node.Name + "{client:client}"
			def_client_node.MethodBodies[constructor.Name] = cons_body
			def_client_node.Constructors = []parser.FuncInfo{constructor}
			def_client_node.Fields = []parser.ArgInfo{parser.GetPointerArg("client", prev_client_node.Name)}
			arg_strings := []string{v.prev_client_name}
			client_name := strings.ToLower(def_client_node.Name)
			for idx, value := range def_client_node.Values {
				if idx == 0 {
					continue
				}
				if val, ok := v.client_names[value]; !ok {
					arg_strings = append(arg_strings, "\""+value+"\"")
				} else {
					arg_strings = append(arg_strings, val)
				}
			}
			v.curBody += client_name + " := " + def_client_node.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
			v.prev_client_name = client_name
		} else {
			// There is no previous client name
			def_client_node := client_nodes[0]
			arg_strings := []string{}
			client_name := strings.ToLower(def_client_node.Name)
			for _, value := range def_client_node.Values {
				if val, ok := v.client_names[value]; !ok {
					arg_strings = append(arg_strings, "\""+value+"\"")
				} else {
					arg_strings = append(arg_strings, val)
				}
			}
			v.curBody += client_name + " := " + def_client_node.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
			v.prev_client_name = client_name
		}
		body += v.curBody
		v.curBody = ""
		client_names[name] = v.prev_client_name
	}

	handler_node := n.ASTServerNodes[0]
	var harg_strings []string
	for _, param := range n.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			harg_strings = append(harg_strings, "\""+ptype.Value+"\"")
		case *InstanceParameter:
			harg_strings = append(harg_strings, client_names[ptype.Name])
		}
	}

	implType := v.impls[handler_node.Name]
	pkg_name := v.pathpkgs[implType.PkgPath]
	import_path := "spec" + strings.ReplaceAll(implType.PkgPath, v.specDir, "")
	imports := []parser.ImportInfo{}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: import_path})
	imports = append(imports, v.client_imports...)

	var import_string string
	for _, imp := range imports {
		import_string += "import \"" + imp.FullName + "\"\n"
	}
	_, err = outf.WriteString(import_string)
	if err != nil {
		v.logger.Fatal(err)
	}

	main_handler_string := "spec_handler := " + pkg_name + "." + handler_node.Constructors[0].Name + "(" + strings.Join(harg_strings, ", ") + ")\n"
	body += main_handler_string
	body += "spec_handler.Entry()\n"
	body += "return nil"
	var arg_strings []string
	for _, arg := range args {
		arg_strings = append(arg_strings, arg.String())
	}
	var ret_strings []string
	for _, arg := range ret_args {
		ret_strings = append(ret_strings, arg.String())
	}
	func_string := "func " + func_name + "(" + strings.Join(arg_strings, ", ") + ") "
	if len(ret_strings) > 1 {
		func_string += "(" + strings.Join(ret_strings, ", ") + ")"
	} else {
		func_string += strings.Join(ret_strings, ", ")
	}
	func_string += " {\n"
	func_string += "\t" + strings.ReplaceAll(body, "\n", "\n\t")
	func_string += "\n}\n"
	_, err = outf.WriteString(func_string + "\n")
	if err != nil {
		v.logger.Fatal(err)
	}
	n.RunFuncName = func_name
	v.runNames = append(v.runNames, func_name)
	v.deployInfo = n.DepInfo
}

func (v *MainVisitor) defaultClientConstructorGeneration(m Modifier) {
	m.AddClientConstructor(v.curClientNode, v.nextClientNode)
	// Use construtor params
	arg_strings := []string{v.prev_client_name}
	client_name := v.getVariableName(v.curClientNode)
	for _, value := range v.curClientNode.Values {
		if val, ok := v.client_names[value]; !ok {
			arg_strings = append(arg_strings, "\""+value+"\"")
		} else {
			arg_strings = append(arg_strings, val)
		}
	}
	v.curBody += client_name + " := " + v.curClientNode.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
	v.prev_client_name = client_name
}

func (v *MainVisitor) VisitTracerModifier(_ Visitor, n *TracerModifier) {
	v.defaultClientConstructorGeneration(n)
}

func (v *MainVisitor) VisitRPCServerModifier(_ Visitor, n *RPCServerModifier) {
	if _, ok := v.added_imports["log"]; !ok {
		v.client_imports = append(v.client_imports, parser.ImportInfo{ImportName: "", FullName: "log"})
		v.added_imports["log"] = true
	}
	n.AddClientConstructor(v.curClientNode, v.nextClientNode)
	clientName := v.getVariableName(v.curClientNode)
	errName := clientName + "_neterr"
	v.curBody += "var " + errName + " error\n"
	v.curBody += "var " + clientName + "_netclient *" + v.curClientNode.Name + "\n"
	v.curBody += "for {\n"
	v.curBody += "\t" + clientName + "_netclient," + errName + " = " + v.curClientNode.Constructors[0].Name + "()\n"
	v.curBody += "\tif " + errName + " == nil{\n"
	v.curBody += "\t\tbreak\n"
	v.curBody += "\t} else {\n"
	v.curBody += "\t\tlog.Println(" + errName + ")\n"
	v.curBody += "\t}\n"
	v.curBody += "}\n"
	v.prev_client_name = clientName + "_netclient"
}

func (v *MainVisitor) VisitWebServerModifier(_ Visitor, n *WebServerModifier) {
	n.AddClientConstructor(v.curClientNode, v.nextClientNode)
	// TODO: Catch error
	clientName := v.getVariableName(v.curClientNode)
	v.curBody += clientName + "_netclient, _ := " + v.curClientNode.Constructors[0].Name + "()\n"
	v.prev_client_name = clientName + "_netclient"
}

func (v *MainVisitor) VisitXTraceModifier(_ Visitor, n *XTraceModifier) {
	v.defaultClientConstructorGeneration(n)
}

func (v *MainVisitor) VisitMetricModifier(_ Visitor, n *MetricModifier) {
	v.defaultClientConstructorGeneration(n)
}

func (v *MainVisitor) VisitClientPoolModifier(_ Visitor, n *ClientPoolModifier) {
	n.AddClientConstructor(v.curClientNode, v.nextClientNode)
	client_name := v.getVariableName(v.curClientNode)
	var body string
	fn_name := client_name + "_fn"
	v.logger.Println("Inside clientpool: Previous client name is", v.prev_client_name)
	body = fn_name + " := func()*" + v.nextClientNode.Name + "{\n\t" + strings.ReplaceAll(v.curBody, "\n", "\n\t")
	body += "return " + v.prev_client_name + "\n}\n"
	v.curBody = body
	var arg_strings []string
	for _, value := range v.curClientNode.Values {
		if val, ok := v.client_names[value]; !ok {
			arg_strings = append(arg_strings, "\""+value+"\"")
		} else {
			arg_strings = append(arg_strings, val)
		}
	}
	arg_strings = append(arg_strings, fn_name)
	v.curBody += client_name + " := " + v.curClientNode.Constructors[0].Name + "(" + strings.Join(arg_strings, ", ") + ")\n"
	v.prev_client_name = client_name
}

func (v *MainVisitor) VisitRetryModifier(_ Visitor, n *RetryModifier) {
	v.defaultClientConstructorGeneration(n)
}

func (v *MainVisitor) VisitCircuitBreakerModifier(_ Visitor, n *CircuitBreakerModifier) {
	v.defaultClientConstructorGeneration(n)
}

func (v *MainVisitor) VisitJaegerNode(_ Visitor, n *JaegerNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.public_ports[v.port] = 14268
	v.public_ports[6832] = 6832
	v.imageName = "jaegertracing/all-in-one:latest"
	v.public_ports[16686] = 16686
	v.public_ports[5775] = 5775
	v.public_ports[6831] = 6831
	v.public_ports[5778] = 5778
}

func (v *MainVisitor) VisitZipkinNode(_ Visitor, n *ZipkinNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.public_ports[v.port] = 9411
	v.imageName = "openzipkin/zipkin"
}

func (v *MainVisitor) VisitXTraceNode(_ Visitor, n *XTraceNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.imageName = "jonathanmace/xtrace-server:latest"
	v.public_ports[4080] = 4080
	v.public_ports[v.port] = 5563
}

func (v *MainVisitor) VisitMemcachedNode(_ Visitor, n *MemcachedNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.imageName = "memcached"
}

func (v *MainVisitor) VisitRedisNode(_ Visitor, n *RedisNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.imageName = "redis"
}

func (v *MainVisitor) VisitMongoDBNode(_ Visitor, n *MongoDBNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.imageName = "mongo"
}

func (v *MainVisitor) VisitRabbitMQNode(_ Visitor, n *RabbitMQNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.hostname = n.DepInfo.Hostname
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.cur_env_vars["RABBITMQ_ERLANG_COOKIE"] = n.Name + "-RABBITMQ"
	v.cur_env_vars["RABBITMQ_DEFAULT_HOST"] = "/"
	v.imageName = "rabbitmq:3.8"
}

func (v *MainVisitor) VisitMySqlDBNode(_ Visitor, n *MySqlDBNode) {
	v.copyEnvVars(n.DepInfo.EnvVars)
	v.isservice = false
	v.address = n.DepInfo.Address
	v.port = n.DepInfo.Port
	v.cur_env_vars[n.Name+"_ADDRESS"] = v.address
	v.cur_env_vars[n.Name+"_PORT"] = strconv.Itoa(v.port)
	v.imageName = "mysql/mysql-server"
}

func (v *MainVisitor) getVariableName(n *ServiceImplInfo) string {
	var variableName string
	node_name := strings.ToLower(n.Name)
	instance_name := strings.ToLower(n.InstanceName)
	variableName = instance_name + "_" + node_name
	return variableName
}
