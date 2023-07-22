package generators

import (
	"log"
)

type LocalServicesInfoCollectorVisitor struct {
	DefaultVisitor
	logger            *log.Logger
	LocalServiceInfos map[string]map[string]string
	curProcName       string
}

func NewLocalServicesInfoCollectorVisitor(logger *log.Logger) *LocalServicesInfoCollectorVisitor {
	return &LocalServicesInfoCollectorVisitor{DefaultVisitor{}, logger, make(map[string]map[string]string), ""}
}

func (v *LocalServicesInfoCollectorVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting LocalServicesInfoCollectorVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending LocalServicesInfoCollectorVisitor visit")
}

func (v *LocalServicesInfoCollectorVisitor) VisitProcessNode(_ Visitor, n *ProcessNode) {
	v.LocalServiceInfos[n.Name] = make(map[string]string)
	v.curProcName = n.Name
	v.DefaultVisitor.VisitProcessNode(v, n)
}

func (v *LocalServicesInfoCollectorVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	last_server_node := n.ASTServerNodes[len(n.ASTServerNodes)-1]
	if _, ok := last_server_node.Methods["Run"]; !ok {
		v.LocalServiceInfos[v.curProcName][last_server_node.InstanceName] = last_server_node.Name
	}
}
