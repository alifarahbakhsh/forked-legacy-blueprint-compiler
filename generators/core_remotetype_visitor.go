package generators

import (
	"log"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type RemoteTypeVisitor struct {
	DefaultVisitor
	logger      *log.Logger
	remoteTypes map[string]*parser.ImplInfo
	pathpkgs    map[string]string
	impls       map[string]*parser.ImplInfo
}

func NewRemoteTypeVisitor(logger *log.Logger, remoteTypes map[string]*parser.ImplInfo, pkgs map[string]string, impls map[string]*parser.ImplInfo) *RemoteTypeVisitor {
	return &RemoteTypeVisitor{DefaultVisitor{}, logger, remoteTypes, pkgs, impls}
}

func (v *RemoteTypeVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting RemoteTypeVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending RemoteTypeVisitor visit")
}

func (v *RemoteTypeVisitor) fixRemoteTypes(node *ServiceImplInfo) {
	rtype := v.impls[node.Name]
	pkgName := v.pathpkgs[rtype.PkgPath]
	has_remote_objs := false
	has_return_remote_objs := false
	var base_imports []parser.ImportInfo
	for name, fInfo := range node.Methods {
		var new_args []parser.ArgInfo
		for _, arg := range fInfo.Args {
			if arg.Type.IsUserDefined() {
				has_remote_objs = true
				arg.Type = parser.PrependPackageName(pkgName, arg.Type)
			}
			new_args = append(new_args, arg)
		}
		var rets []parser.ArgInfo
		for _, arg := range fInfo.Return {
			if arg.Type.IsUserDefined() {
				arg.Type = parser.PrependPackageName(pkgName, arg.Type)
				has_return_remote_objs = true
			}
			rets = append(rets, arg)
		}
		fInfo.Args = new_args
		fInfo.Return = rets
		node.Methods[name] = fInfo
	}
	if has_remote_objs || has_return_remote_objs {
		base_imports = append(base_imports, parser.ImportInfo{ImportName: "", FullName: "spec/" + pkgName})
	}
	node.BaseImports = base_imports
	node.HasUserDefinedObjs = has_remote_objs
	node.HasReturnDefinedObjs = has_return_remote_objs
}

func (v *RemoteTypeVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.logger.Println("Fixing Remote Type for", n.Name)
	server_node := n.ASTServerNodes[len(n.ASTServerNodes)-1]
	v.fixRemoteTypes(server_node)
}
