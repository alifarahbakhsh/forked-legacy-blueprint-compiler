package generators

import (
	"log"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type GenerateClientSourceCodeVisitor struct {
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
	services    map[string]*parser.ServiceInfo
	DepGraph    *DependencyGraph
}

func NewGenerateClientSourceCodeVisitor(logger *log.Logger, modregistry *ModifierRegistry, appName string, outdir string, remoteTypes map[string]*parser.ImplInfo, cinfos map[string]*ClientInfo, pkgs map[string]string, impls map[string]*parser.ImplInfo, specDir string, enums map[string]*parser.EnumInfo, services map[string]*parser.ServiceInfo, depGraph *DependencyGraph) *GenerateClientSourceCodeVisitor {
	return &GenerateClientSourceCodeVisitor{DefaultVisitor{}, logger, modregistry, appName, outdir, remoteTypes, cinfos, pkgs, impls, specDir, enums, services, depGraph}
}

func (v *GenerateClientSourceCodeVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting GeneratingClientSourceCodeVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending GeneratingClientSourceCodeVisitor visit")
}

func (v *GenerateClientSourceCodeVisitor) getPkgPath(name string) string {
	if rtype, ok := v.impls[name]; ok {
		return rtype.PkgPath
	} else if rtype1, ok := v.services[name]; ok {
		return rtype1.PkgPath
	}

	return ""
}

func (v *GenerateClientSourceCodeVisitor) fixDefaultClient(node *ServiceImplInfo) {
	pkgpath := v.getPkgPath(node.BaseName)
	if pkgpath != "" {
		pkgName := v.pathpkgs[pkgpath]
		for i, fInfo := range node.Methods {
			for j, arg := range fInfo.Args {
				// TODO: This won't work for list + map types. That requires special handling.
				if _, ok2 := v.remoteTypes[arg.Type.String()]; ok2 {
					arg.Type.Detail.UserType = pkgName + "." + arg.Type.Detail.UserType
					fInfo.Args[j] = arg
				}
			}
			for j, retarg := range fInfo.Return {
				// TODO: This won't work for list + map types. That requires special handling.
				if _, ok2 := v.remoteTypes[retarg.Type.String()]; ok2 {
					retarg.Type.Detail.UserType = pkgName + "." + retarg.Type.Detail.UserType
					fInfo.Return[j] = retarg
				}
			}
			node.Methods[i] = fInfo
		}
	} else {
		v.logger.Println("Couldn't fix default client for", node.BaseName)
	}
}

func (v *GenerateClientSourceCodeVisitor) getClientNodes(name string, dependencies map[string]Dependency, modifierClientNodes map[string][]*ServiceImplInfo, paramClientNodes map[string][]*ServiceImplInfo) {
	instances := make(map[string]bool)
	topo_ordering := v.DepGraph.Order[name]
	for _, dependency := range topo_ordering {
		v.logger.Println("Applying client modifiers for", dependency)
		if cinfo, ok := v.cinfos[dependency]; !ok {
			v.logger.Fatal("Could not find instance with name", dependency)
		} else {
			modifiers := make([]Modifier, len(cinfo.ClientModifiers))
			copy(modifiers, cinfo.ClientModifiers)
			var client_nodes []*ServiceImplInfo
			client_nodes = append(client_nodes, copyServiceImplInfo(cinfo.ClientNode))
			// Override modifiers
			var overriden_modifiers []Modifier
			if dependency_info, ok := dependencies[dependency]; ok {
				overriden_modifiers = dependency_info.ClientModifiers
			}
			for _, modifier := range overriden_modifiers {
				found := false
				for idx, def_modifier := range modifiers {
					if def_modifier.GetName() == modifier.GetName() {
						modifiers[idx] = modifier
						found = true
					}
				}
				if !found {
					modifiers = append(modifiers, modifier)
				}
			}
			prev_client_node := client_nodes[len(client_nodes)-1]
			for _, modifier := range v.modregistry.OrderModifiers(modifiers) {
				v.logger.Println("Applying modifier", modifier.GetName())
				if !modifier.IsSourceCodeModifier() {
					// Only need to apply source code modifiers in this visit
					continue
				}
				var new_client_node *ServiceImplInfo
				var err error
				if cinfo.IsService {
					new_client_node, err = modifier.ModifyClient(prev_client_node)
					if err != nil {
						v.logger.Fatal(err)
					}
				} else if cinfo.IsQueue {
					new_client_node, err = modifier.ModifyQueue(prev_client_node)
					if err != nil {
						v.logger.Fatal(err)
					}
				}
				// TODO: Add ModifyComponent
				if new_client_node != nil {
					new_client_node.PluginName = modifier.GetPluginName()
					new_client_node.ModifierParams = modifier.GetParams()
					for _, mod_param := range new_client_node.ModifierParams {
						switch ptype := mod_param.(type) {
						case *ValueParameter:
							new_client_node.Values = append(new_client_node.Values, ptype.Value)
						case *InstanceParameter:
							new_client_node.Values = append(new_client_node.Values, ptype.Name)
							instances[ptype.Name] = true
						}
					}
					new_client_node.BaseImports = prev_client_node.BaseImports
					new_client_node.HasUserDefinedObjs = prev_client_node.HasUserDefinedObjs
					new_client_node.HasReturnDefinedObjs = prev_client_node.HasReturnDefinedObjs
					new_client_node.ModifierNode = modifier
					client_nodes = append(client_nodes, new_client_node)
					prev_client_node = new_client_node
				}
			}
			if cinfo.IsService {
				v.fixDefaultClient(client_nodes[0])
			} else if cinfo.IsQueue {
				client_nodes = append(client_nodes, cinfo.FinalClientNode)
			}
			paramClientNodes[dependency] = client_nodes
		}
	}

	for name, _ := range instances {
		if _, ok := modifierClientNodes[name]; ok {
			// Already saw this modifier before
			continue
		}
		if cinfo, ok := v.cinfos[name]; !ok {
			v.logger.Fatal("Could not find instance with name ", name)
		} else {
			var client_nodes []*ServiceImplInfo
			client_nodes = append(client_nodes, copyServiceImplInfo(cinfo.ClientNode))
			modifierClientNodes[name] = client_nodes
		}
	}
}

func (v *GenerateClientSourceCodeVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.logger.Println("Generating Client Source Code Modifications for", n.Name)
	v.logger.Println("Dependency Order:", v.DepGraph.Order[n.Name])
	v.getClientNodes(n.Name, n.GetDependencies(), n.ModifierClientNodes, n.ParamClientNodes)
}

func (v *GenerateClientSourceCodeVisitor) VisitQueueServiceNode(_ Visitor, n *QueueServiceNode) {
	v.logger.Println("Generating Client Source Code Modifications for", n.Name)
	v.getClientNodes(n.Name, n.GetDependencies(), n.ModifierClientNodes, n.ParamClientNodes)
}
