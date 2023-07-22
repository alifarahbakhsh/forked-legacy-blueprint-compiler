package generators

import (
	"errors"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type DockerContainerNode struct {
	Name     string
	Children []Node
	Options  map[string]interface{}
}

type AnsibleContainerNode struct {
	Name     string
	Children []Node
}

type KubernetesContainerNode struct {
	Name     string
	Children []Node
}

type NoOpContainerNode struct {
	Name     string
	Children []Node
}

func (n *DockerContainerNode) Accept(v Visitor) {
	v.VisitDockerContainerNode(v, n)
}

func (n *DockerContainerNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (n *AnsibleContainerNode) Accept(v Visitor) {
	v.VisitAnsibleContainerNode(v, n)
}

func (n *AnsibleContainerNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (n *KubernetesContainerNode) Accept(v Visitor) {
	v.VisitKubernetesContainerNode(v, n)
}

func (n *KubernetesContainerNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func (n *NoOpContainerNode) Accept(v Visitor) {
	v.VisitNoOpContainerNode(v, n)
}

func (n *NoOpContainerNode) GetNodes(nodeType string) []Node {
	var nodes []Node
	if getType(n) == nodeType {
		nodes = append(nodes, n)
	}
	for _, child := range n.Children {
		nodes = append(nodes, child.GetNodes(nodeType)...)
	}
	return nodes
}

func GetDeployerNode(deployer parser.ModifierNode, Name string, Children []Node) (Container, error) {
	var deployer_name string
	deployer_opts := make(map[string]interface{})
	for _, arg := range deployer.ModifierParams {
		if arg.KeywordName == "framework" {
			deployer_name = arg.Value
		} else {
			deployer_opts[arg.KeywordName] = arg.Value
		}
	}

	if deployer_name == "" {
		return nil, errors.New("No deployer framework specified")
	} else {
		switch deployer_name {
		case "'docker'":
			return &DockerContainerNode{Name: Name, Children: Children, Options: deployer_opts}, nil
		case "'kubernetes'":
			dockerNode := DockerContainerNode{Name: Name, Children: Children, Options: deployer_opts}
			return &KubernetesContainerNode{Name: Name, Children: []Node{&dockerNode}}, nil
		case "'ansible'":
			// dockerNode := DockerContainerNode{Name: Name + "_docker", Children: Children, Options: deployer_opts}
			dockerNode := DockerContainerNode{Name: Name, Children: Children, Options: deployer_opts}
			return &AnsibleContainerNode{Name: Name, Children: []Node{&dockerNode}}, nil
		case "'noop'":
			return &NoOpContainerNode{Name: Name, Children: Children}, nil
		default:
			return nil, errors.New("Unsupported deployer framework " + deployer_name)
		}
	}
}
