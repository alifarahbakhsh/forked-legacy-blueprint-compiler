package generators

import (
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type LocalMetricNode struct {
	Name            string
	Params          []Parameter
	ClientModifiers []Modifier
	ServerModifiers []Modifier
}

func (n *LocalMetricNode) Accept(v Visitor) {
	v.VisitLocalMetricNode(v, n)
}

func (n *LocalMetricNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
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

func GenerateLocalMetricNode(node parser.DetailNode) Node {
	var params []Parameter
	var cmodifiers []Modifier
	var smodifiers []Modifier
	for _, arg := range node.Arguments {
		params = append(params, convert_argument_node(arg))
	}
	for _, modifier := range node.ClientModifiers {
		cmodifiers = append(cmodifiers, convert_modifier_node(modifier))
	}
	for _, modifier := range node.ServerModifiers {
		smodifiers = append(smodifiers, convert_modifier_node(modifier))
	}

	return &LocalMetricNode{Name: node.Name, Params: params, ClientModifiers: cmodifiers, ServerModifiers: smodifiers}
}
