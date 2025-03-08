package generators

import (
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type MetricModifier struct {
	*NoOpSourceCodeModifier
	Params []Parameter
}

func (m *MetricModifier) Accept(v Visitor) {
	v.VisitMetricModifier(v, m)
}

func (m *MetricModifier) GetParams() []Parameter {
	return m.Params
}

func (n *MetricModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *MetricModifier) GetName() string {
	return "MetricModifier"
}

func (m *MetricModifier) GetPluginName() string {
	return "Metric"
}

func GenerateMetricModifier(node parser.ModifierNode) Modifier {
	return &MetricModifier{NewNoOpSourceCodeModifier(), get_params(node)}
}
