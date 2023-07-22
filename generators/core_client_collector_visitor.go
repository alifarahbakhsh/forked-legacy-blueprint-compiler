package generators

import (
	"fmt"
	"log"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type ClientInfo struct {
	ClientModifiers []Modifier
	ClientNode      *ServiceImplInfo
	FinalClientNode *ServiceImplInfo
	IsService       bool
	IsQueue         bool
	IsComponent     bool
}

type ClientCollectorVisitor struct {
	DefaultVisitor
	logger             *log.Logger
	DefaultClientInfos map[string]*ClientInfo
	impls              map[string]*parser.ImplInfo
	pathpkgs           map[string]string
	specDir            string
	remoteTypes        map[string]*parser.ImplInfo
	services           map[string]*parser.ServiceInfo
}

func (v *ClientCollectorVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting ClientCollectorVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending ClientCollectorVisitor visit")
}

func (v *ClientCollectorVisitor) generateClientNode(server_node *ServiceImplInfo) *ServiceImplInfo {
	name := server_node.Name + "Client"
	receiverName := "this"
	funcs := copyMap(server_node.Methods)
	methods := make(map[string]string)
	for key, fInfo := range funcs {
		var arg_names []string
		for _, arg := range fInfo.Args {
			arg_names = append(arg_names, arg.Name)
		}
		var retarg_names []string
		var ret_string_names []string
		for idx, _ := range fInfo.Return {
			retarg_name := fmt.Sprintf("ret%d", idx)
			retarg_names = append(retarg_names, retarg_name)
			ret_string_names = append(ret_string_names, retarg_name)
		}
		body := "return " + receiverName + ".client." + fInfo.Name + "(" + strings.Join(arg_names, ", ") + ")"
		methods[key] = body
	}
	imports := []parser.ImportInfo{parser.ImportInfo{ImportName: "", FullName: "context"}}
	v.logger.Println(server_node.BaseImports)
	return &ServiceImplInfo{Name: name, ReceiverName: receiverName, Methods: funcs, MethodBodies: methods, BaseName: server_node.BaseName, Imports: imports, InstanceName: server_node.InstanceName, HasUserDefinedObjs: server_node.HasUserDefinedObjs, BaseImports: server_node.BaseImports, HasReturnDefinedObjs: server_node.HasReturnDefinedObjs, PluginName: "Blueprint Core"}
}

func (v *ClientCollectorVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)

	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: v.generateClientNode(n.ASTServerNodes[0]), IsService: true}
}

func (v *ClientCollectorVisitor) VisitLoadBalancerNode(_ Visitor, n *LoadBalancerNode) {
	v.logger.Println("LoadBalancer doesn't allow any modifiers.", n.Name)
	sinfo := v.services[n.BaseTypeName]
	n.GenerateClientNode(sinfo)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientNode: n.ASTNodes[0], IsService: true}
}

func (v *ClientCollectorVisitor) VisitJaegerNode(_ Visitor, n *JaegerNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitZipkinNode(_ Visitor, n *ZipkinNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitXTraceNode(_ Visitor, n *XTraceNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitMemcachedNode(_ Visitor, n *MemcachedNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitRedisNode(_ Visitor, n *RedisNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitMongoDBNode(_ Visitor, n *MongoDBNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func (v *ClientCollectorVisitor) VisitRabbitMQNode(_ Visitor, n *RabbitMQNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsQueue: true, FinalClientNode: n.ASTNodes[1]}
}

func (v *ClientCollectorVisitor) VisitMySqlDBNode(_ Visitor, n *MySqlDBNode) {
	v.logger.Println("Finding default modifiers for service", n.Name)
	all_modifiers := make([]Modifier, len(n.ServerModifiers))
	copy(all_modifiers, n.ServerModifiers)
	all_modifiers = append(all_modifiers, n.ClientModifiers...)
	impl_info := v.impls[n.TypeName]
	n.GenerateClientNode(impl_info)
	v.DefaultClientInfos[n.Name] = &ClientInfo{ClientModifiers: all_modifiers, ClientNode: n.ASTNodes[0], IsComponent: true}
}

func NewClientCollectorVisitor(logger *log.Logger, impls map[string]*parser.ImplInfo, pathpkgs map[string]string, specDir string, remoteTypes map[string]*parser.ImplInfo, services map[string]*parser.ServiceInfo) *ClientCollectorVisitor {
	return &ClientCollectorVisitor{DefaultVisitor{}, logger, make(map[string]*ClientInfo), impls, pathpkgs, specDir, remoteTypes, services}
}
