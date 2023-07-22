package generators

import (
	"log"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/netgen"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type GenerateSourceCodeVisitor struct {
	DefaultVisitor
	logger      *log.Logger
	modregistry *ModifierRegistry
	appName     string
	outdir      string
	remoteTypes map[string]*parser.ImplInfo
	cinfos      map[string]*ClientInfo
	pathpkgs    map[string]string
	impls       map[string]*parser.ImplInfo
	specDir     string
	enums       map[string]*parser.EnumInfo
	Frameworks  map[string]netgen.NetworkGenerator
}

func NewGenerateSourceCodeVisitor(logger *log.Logger, modregistry *ModifierRegistry, appName string, outdir string, remoteTypes map[string]*parser.ImplInfo, cinfos map[string]*ClientInfo, pkgs map[string]string, impls map[string]*parser.ImplInfo, specDir string, enums map[string]*parser.EnumInfo) *GenerateSourceCodeVisitor {
	return &GenerateSourceCodeVisitor{DefaultVisitor{}, logger, modregistry, appName, outdir, remoteTypes, cinfos, pkgs, impls, specDir, enums, make(map[string]netgen.NetworkGenerator)}
}

func (v *GenerateSourceCodeVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting GeneratingSourceCodeVisitor visit")
	rpc_nodes := n.GetNodes("RPCServerModifier")
	seen_frameworks := make(map[string]bool)
	for _, node := range rpc_nodes {
		rpc_node := node.(*RPCServerModifier)
		framework, netgenerator, err := rpc_node.GetFrameworkInfo()
		if err != nil {
			v.logger.Fatal(err)
		}
		if _, ok := seen_frameworks[framework]; !ok {
			v.logger.Println("Converting remote types for NetworkGenerator framework", framework)
			netgenerator.SetAppName(v.appName)
			err := netgenerator.ConvertRemoteTypes(v.remoteTypes)
			if err != nil {
				v.logger.Fatal(err)
			}
			err = netgenerator.ConvertEnumTypes(v.enums)
			if err != nil {
				v.logger.Fatal(err)
			}
			v.Frameworks[framework] = netgenerator
		}
		seen_frameworks[framework] = true
	}
	v.DefaultVisitor.VisitMillenialNode(v, n)
	generated_files := make(map[string]bool)
	for _, node := range rpc_nodes {
		rpc_node := node.(*RPCServerModifier)
		framework, netgenerator, err := rpc_node.GetFrameworkInfo()
		if err != nil {
			v.logger.Fatal(err)
		}
		if _, ok := generated_files[framework]; !ok {
			v.logger.Println("Generating files for NetworkGenerator framework", framework)
			err := netgenerator.GenerateFiles(v.outdir)
			if err != nil {
				v.logger.Fatal(err)
			}
			generated_files[framework] = true
		}
	}
	v.logger.Println("Ending GeneratingSourceCodeVisitor visit")
}

func (v *GenerateSourceCodeVisitor) modifyDefaultServiceNode(node *ServiceImplInfo) {
	rtype := v.impls[node.Name]
	pkgName := v.pathpkgs[rtype.PkgPath]
	fields := []parser.ArgInfo{parser.GetPointerArg("service", pkgName+"."+node.Name)}
	node.Fields = fields
	for name, fInfo := range node.Methods {
		var arg_names []string
		for _, arg := range fInfo.Args {
			arg_names = append(arg_names, arg.Name)
		}
		body := "return " + node.ReceiverName + ".service." + name + "(" + strings.Join(arg_names, ", ") + ")"
		node.MethodBodies[name] = body
	}
	import_path := "spec" + strings.ReplaceAll(rtype.PkgPath, v.specDir, "")
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: import_path}}
	imports = append(imports, parser.ImportInfo{ImportName: "", FullName: "context"})
	node.Imports = imports
	con_args := []parser.ArgInfo{parser.GetPointerArg("handler", pkgName+"."+node.Name)}
	con_rets := []parser.ArgInfo{parser.GetPointerArg("", node.Name)}
	con_name := "New" + node.Name
	constructor := parser.FuncInfo{Name: con_name, Args: con_args, Return: con_rets}
	node.Constructors = []parser.FuncInfo{constructor}
	node.MethodBodies[con_name] = "return &" + node.Name + "{service:handler}"
	node.PluginName = "Blueprint Core"
}

func (v *GenerateSourceCodeVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	// Generate Default Client Node
	v.logger.Println("Generatring Source Code Modifications for", n.Name)
	prev_server_node := n.ASTServerNodes[len(n.ASTServerNodes)-1]

	// Modifier instances that also need to be captured and stored in the
	instances := make(map[string]bool)
	v.modifyDefaultServiceNode(prev_server_node)
	// Order Modifiers: Network modifier has to go last
	v.logger.Println("Modifying service nodes")
	for _, modifier := range v.modregistry.OrderModifiers(n.ServerModifiers) {
		if !modifier.IsSourceCodeModifier() {
			// Only need to apply the source code modifiers in this visit
			continue
		}
		v.logger.Println("Applying Modifier:", modifier.GetName())
		new_server_node, err := modifier.ModifyServer(prev_server_node)
		if err != nil {
			v.logger.Fatal(err)
		}
		if new_server_node != nil {
			new_server_node.PluginName = modifier.GetPluginName()
			new_server_node.ModifierParams = modifier.GetParams()
			for _, param := range new_server_node.ModifierParams {
				switch ptype := param.(type) {
				case *ValueParameter:
					new_server_node.Values = append(new_server_node.Values, ptype.Value)
				case *InstanceParameter:
					new_server_node.Values = append(new_server_node.Values, ptype.Name)
					instances[ptype.Name] = true
				}
			}
			new_server_node.HasUserDefinedObjs = prev_server_node.HasUserDefinedObjs
			new_server_node.HasReturnDefinedObjs = prev_server_node.HasReturnDefinedObjs
			n.ASTServerNodes = append(n.ASTServerNodes, new_server_node)
			prev_server_node = new_server_node
		}
	}

	for name, _ := range instances {
		if cinfo, ok := v.cinfos[name]; !ok {
			v.logger.Fatal("Could not find instance with name ", name)
		} else {
			var client_nodes []*ServiceImplInfo
			client_nodes = append(client_nodes, copyServiceImplInfo(cinfo.ClientNode))
			n.ModifierClientNodes[name] = client_nodes
		}
	}
}
