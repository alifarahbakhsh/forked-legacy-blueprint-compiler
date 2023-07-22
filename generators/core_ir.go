package generators

import (
	"log"
	"reflect"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

func getType(myVar interface{}) string {
	t := reflect.TypeOf(myVar)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

type Node interface {
	Accept(v Visitor)
	GetNodes(nodeType string) []Node
}

type Container interface {
	Node
}

type MillenialNode struct {
	Children []Node
}

func (n *MillenialNode) Accept(v Visitor) {
	v.VisitMillenialNode(v, n)
}

func (n *MillenialNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

type ProcessNode struct {
	Name     string
	Children []Node
}

func (n *ProcessNode) Accept(v Visitor) {
	v.VisitProcessNode(v, n)
}

func (n *ProcessNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

type Parameter interface {
	Node
	GetKeywordName() string
	GetValue() string
}

type InstanceParameter struct {
	KeywordName     string
	Name            string
	ClientModifiers []Modifier
}

func (p *InstanceParameter) Accept(v Visitor) {
	v.VisitInstanceParameter(v, p)
}

func (n *InstanceParameter) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.ClientModifiers {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (p *InstanceParameter) GetKeywordName() string {
	return p.KeywordName
}

func (p *InstanceParameter) GetValue() string {
	return p.Name
}

type ValueParameter struct {
	KeywordName string
	Value       string
}

func (p *ValueParameter) Accept(v Visitor) {
	v.VisitValueParameter(v, p)
}

func (n *ValueParameter) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	return nodes
}

func (p *ValueParameter) GetKeywordName() string {
	return p.KeywordName
}

func (p *ValueParameter) GetValue() string {
	return p.Value
}

type Modifier interface {
	Node
	GetName() string
	GetPluginName() string
	IsSourceCodeModifier() bool
	IsDeployerModifier() bool
	ModifyServer(prev_node *ServiceImplInfo) (*ServiceImplInfo, error)
	ModifyClient(prev_node *ServiceImplInfo) (*ServiceImplInfo, error)
	ModifyQueue(prev_node *ServiceImplInfo) (*ServiceImplInfo, error)
	ModifyDeployInfo(depInfo *deploy.DeployInfo) error
	AddClientConstructor(node *ServiceImplInfo, next_node *ServiceImplInfo)
	GetParams() []Parameter
}

type ServiceImplInfo struct {
	Name                 string
	ReceiverName         string
	InstanceName         string
	Methods              map[string]parser.FuncInfo
	MethodBodies         map[string]string
	BaseName             string
	Constructors         []parser.FuncInfo
	Imports              []parser.ImportInfo
	Fields               []parser.ArgInfo
	ModifierParams       []Parameter
	Values               []string
	NextNodeMethodArgs   []parser.ArgInfo    // Only used for client so far
	NextNodeMethodReturn []parser.ArgInfo    // Only used for client so far
	Structs              []parser.StructInfo // Captures any dependent objects!
	ModifierNode         Modifier            // ModifierNode used to generate this ServiceImplInfo. This is needed for client-side constructor generation
	HasUserDefinedObjs   bool
	HasReturnDefinedObjs bool
	BaseImports          []parser.ImportInfo
	PluginName           string
}

type ServiceNode struct {
	Name                string
	Children            []Node
	Type                string
	AbstractType        string
	Params              []Parameter
	ClientModifiers     []Modifier
	ServerModifiers     []Modifier
	ASTServerNodes      []*ServiceImplInfo
	ParamClientNodes    map[string][]*ServiceImplInfo
	ModifierClientNodes map[string][]*ServiceImplInfo
	RunFuncName         string
	DepInfo             *deploy.DeployInfo
}

type Dependency struct {
	ClientModifiers []Modifier
	InstanceName    string
}

func (n *ServiceNode) GetDependencies() map[string]Dependency {
	dependencies := make(map[string]Dependency)
	for _, param := range n.Params {
		switch pt := param.(type) {
		case *InstanceParameter:
			dependencies[pt.Name] = Dependency{InstanceName: pt.Name, ClientModifiers: pt.ClientModifiers}
		case *ValueParameter:
			continue
		}
	}
	return dependencies
}

type Initializable interface {
	GetInitializer() string
}

type FuncServiceNode struct {
	ServiceNode
}

type QueueServiceNode struct {
	ServiceNode
}

func (n *FuncServiceNode) Accept(v Visitor) {
	v.VisitFuncServiceNode(v, n)
}

func (n *FuncServiceNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
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

func (n *QueueServiceNode) Accept(v Visitor) {
	v.VisitQueueServiceNode(v, n)
}

func (n *QueueServiceNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
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

var modreg *ModifierRegistry

type Generator struct {
	config       *parser.Config
	logger       *log.Logger
	RootNode     *MillenialNode
	Registry     *IRExtensionRegistry
	ModRegistry  *ModifierRegistry
	ServiceImpls map[string]*parser.ImplInfo
}

func NewGenerator(config *parser.Config, logger *log.Logger, serviceImpls map[string]*parser.ImplInfo, modregistry *ModifierRegistry) *Generator {
	registry := InitIRRegistry(logger)
	modreg = modregistry
	return &Generator{config: config, logger: logger, RootNode: nil, Registry: registry, ModRegistry: modreg, ServiceImpls: serviceImpls}
}

func convert_argument_node(node parser.ArgumentNode) Parameter {
	if node.IsService {
		var modifiers []Modifier
		for _, modifier := range node.ClientModifiers {
			modifiers = append(modifiers, convert_modifier_node(modifier))
		}
		return &InstanceParameter{Name: node.Name, KeywordName: node.KeywordName, ClientModifiers: modifiers}
	} else {
		return &ValueParameter{KeywordName: node.KeywordName, Value: strings.ReplaceAll(node.Value, "'", "")}
	}
}

func convert_modifier_node(node parser.ModifierNode) Modifier {
	return modreg.GetModifier(node)
}

func (g *Generator) convert_detail_node(node parser.DetailNode) (Node, *parser.ModifierNode) {
	if node.AbsType == "Process" {
		var children []Node
		var deployer_node parser.ModifierNode
		for _, child := range node.Children {
			child_node, deployer := g.convert_detail_node(child)
			children = append(children, child_node)
			if deployer != nil {
				deployer_node = *deployer
			}
		}
		return &ProcessNode{Name: node.Name, Children: children}, &deployer_node
	} else if node.AbsType == "Service" || node.AbsType == "QueueService" || strings.HasSuffix(node.AbsType, "Service") {
		var params []Parameter
		var cmodifiers []Modifier
		var smodifiers []Modifier
		var deployer_node parser.ModifierNode
		for _, arg := range node.Arguments {
			params = append(params, convert_argument_node(arg))
		}
		for _, modifier := range node.ClientModifiers {
			cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
		}
		for _, modifier := range node.ServerModifiers {
			if modifier.ModifierType == "Deployer" {
				deployer_node = modifier
				continue
			}
			smodifiers = append(smodifiers, convert_modifier_node(modifier))
		}
		var serverASTNodes []*ServiceImplInfo
		if implInfo, ok := g.ServiceImpls[node.Type]; ok {
			constructors := make([]parser.FuncInfo, len(implInfo.ConstructorInfos))
			copy(constructors, implInfo.ConstructorInfos)
			info := &ServiceImplInfo{Name: node.Type, ReceiverName: "this", Methods: copyMap(implInfo.Methods), MethodBodies: make(map[string]string), BaseName: node.Type, Constructors: constructors, InstanceName: node.Name}
			serverASTNodes = append(serverASTNodes, info)
			g.logger.Println("Found impl info for", node.Type, "for service instance", node.Name)
		} else {
			g.logger.Fatal("Implementation info not found for", node.Type)
		}
		snode := ServiceNode{Name: node.Name, Type: node.Type, Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers, ASTServerNodes: serverASTNodes, ParamClientNodes: make(map[string][]*ServiceImplInfo), ModifierClientNodes: make(map[string][]*ServiceImplInfo), DepInfo: deploy.NewDeployInfo(), AbstractType: node.AbsType}
		if strings.HasSuffix(node.AbsType, "QueueService") {
			return &QueueServiceNode{snode}, &deployer_node
		} else if strings.HasSuffix(node.AbsType, "Service") {
			return &FuncServiceNode{snode}, &deployer_node
		}
	} else {
		// These are component specific nodes!
		var deployer_node *parser.ModifierNode
		deployer_idx := -1
		for idx, modifier := range node.ServerModifiers {
			if modifier.ModifierType == "Deployer" {
				deployer_node = &modifier
				deployer_idx = idx
				break
			}
		}
		if deployer_idx == -1 {
			g.logger.Fatal("No deployer attached to service instance ", node.Name)
		}
		serverModifiers := node.ServerModifiers[:deployer_idx]
		serverModifiers = append(serverModifiers, node.ServerModifiers[deployer_idx+1:]...)
		node.ServerModifiers = serverModifiers
		return g.Registry.GetNode(node), deployer_node
	}
	return nil, nil
}

func (g *Generator) create_container_node(node parser.ContainerNode) Node {
	var children []Node
	var deployer_node *parser.ModifierNode
	for _, child := range node.Children {
		child_node, deployer := g.convert_detail_node(child)
		children = append(children, child_node)
		if deployer != nil {
			if deployer_node != nil {
				g.logger.Fatal("Multiple deployer options provided")
			} else {
				deployer_node = deployer
			}
		}
	}
	if deployer_node == nil {
		g.logger.Fatal("No deployer option specified")
	}
	ctrNode, err := GetDeployerNode(*deployer_node, node.Name, children)
	if err != nil {
		g.logger.Fatal(err)
	}
	return ctrNode
}

func (g *Generator) ConvertSerializedRep(node *parser.MillenialNode) {
	var children []Node
	for _, child := range node.Children {
		ctr_node := g.create_container_node(child)
		children = append(children, ctr_node)
	}
	rootNode := MillenialNode{Children: children}
	g.RootNode = &rootNode
}
