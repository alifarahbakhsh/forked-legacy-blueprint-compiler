package generators

import (
	"log"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type DependencyGraph struct {
	edges   map[string][]string
	visited map[string]bool
	order   []string
	Order   map[string][]string // Stores the ordering of
}

func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{edges: make(map[string][]string), visited: make(map[string]bool), Order: make(map[string][]string)}
}

func (g *DependencyGraph) AddEdge(to string, from string) {
	if v, ok := g.edges[from]; !ok {
		g.edges[from] = []string{to}
	} else {
		g.edges[from] = append(v, to)
	}
}

func (g *DependencyGraph) topoSortHelper(root string, is_root bool, localServices map[string]bool) {
	if v, ok := g.visited[root]; ok && v {
		return
	}
	g.visited[root] = true
	if is_root {
		if v, ok := g.edges[root]; ok {
			for _, edge := range v {
				g.topoSortHelper(edge, false, localServices)
			}
		}
	}
	// Don't append the starting node
	if !is_root {
		g.order = append(g.order, root)
	}
}

func (g *DependencyGraph) TopoSort(coLocatedServices map[string][]string) {
	// Run a topo sort from all nodes
	for node := range g.edges {
		localServices := make(map[string]bool)
		coLocated := coLocatedServices[node]
		for _, service := range coLocated {
			localServices[service] = true
		}
		g.order = []string{}
		g.topoSortHelper(node, true, localServices)
		g.Order[node] = g.order
		g.visited = make(map[string]bool)
	}
}

func (g *DependencyGraph) String() string {
	out_string := ""
	for src, dsts := range g.edges {
		out_string += src + " -> {" + strings.Join(dsts, ", ") + "}\n"
	}
	return out_string
}

type DependencyGraphVisitor struct {
	DefaultVisitor
	logger   *log.Logger
	impls    map[string]*parser.ImplInfo
	services map[string]*parser.ServiceInfo
	DepGraph *DependencyGraph
}

func (v *DependencyGraphVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Building Dependency Graph")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Finished building the Dependency Graph")
}

func (v *DependencyGraphVisitor) addDependencies(deps map[string]Dependency, name string) {
	for _, dependency := range deps {
		v.DepGraph.AddEdge(dependency.InstanceName, name)
	}
}

func (v *DependencyGraphVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.addDependencies(n.GetDependencies(), n.Name)
}

func (v *DependencyGraphVisitor) VisitQueueServiceNode(_ Visitor, n *QueueServiceNode) {
	v.addDependencies(n.GetDependencies(), n.Name)
}

func (v *DependencyGraphVisitor) VisitLoadBalancerNode(_ Visitor, n *LoadBalancerNode) {
	v.addDependencies(n.GetDependencies(), n.Name)
}

func NewDependencyGraphVisitor(logger *log.Logger, impls map[string]*parser.ImplInfo, services map[string]*parser.ServiceInfo) *DependencyGraphVisitor {
	graph := NewDependencyGraph()
	return &DependencyGraphVisitor{DefaultVisitor{}, logger, impls, services, graph}
}
