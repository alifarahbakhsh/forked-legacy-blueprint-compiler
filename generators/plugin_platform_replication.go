package generators

import (
	"strconv"

	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/generators/deploy"
	"github.com/alifarahbakhsh/forked-legacy-blueprint-compiler/parser"
)

type PlatformReplicationModifier struct {
	*NoOpDeployerModifier
	Params []Parameter
}

func (m *PlatformReplicationModifier) Accept(v Visitor) {
	v.VisitPlaformReplicationModifier(v, m)
}

func (m *PlatformReplicationModifier) GetParams() []Parameter {
	return m.Params
}

func (n *PlatformReplicationModifier) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Params {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (m *PlatformReplicationModifier) GetName() string {
	return "PlatformReplication"
}

func (m *PlatformReplicationModifier) GetPluginName() string {
	return "PlatformReplication"
}

func (m *PlatformReplicationModifier) ModifyDeployInfo(depInfo *deploy.DeployInfo) error {
	arg_map := make(map[string]string)
	for _, param := range m.Params {
		switch ptype := param.(type) {
		case *ValueParameter:
			arg_map[ptype.KeywordName] = ptype.Value
		}
	}
	num_replicas, err := strconv.Atoi(arg_map["num_replicas"])
	depInfo.NumReplicas = num_replicas
	if err != nil {
		return err
	}
	return nil
}

func GeneratePlaformReplicationModifier(node parser.ModifierNode) Modifier {
	return &PlatformReplicationModifier{NewNoOpDeployerModifier(), get_params(node)}
}
