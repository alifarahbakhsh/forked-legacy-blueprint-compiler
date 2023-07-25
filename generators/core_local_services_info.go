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

type CoLocatedServiceInfosVisitor struct {
	DefaultVisitor
	logger            *log.Logger
	CoLocatedServices map[string][]string
	curProcName       string
	procServices      map[string][]string
}

func NewCoLocatedServiceInfosVisitor(logger *log.Logger) *CoLocatedServiceInfosVisitor {
	return &CoLocatedServiceInfosVisitor{DefaultVisitor{}, logger, make(map[string][]string), "", make(map[string][]string)}
}

func (v *CoLocatedServiceInfosVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting CoLocatedServiceInfosVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending CoLocatedServiceInfosVisitor visit")
	for _, names := range v.procServices {
		for _, name := range names {
			v.CoLocatedServices[name] = []string{}
			for _, name2 := range names {
				if name != name2 {
					v.CoLocatedServices[name] = append(v.CoLocatedServices[name], name2)
				}
			}
		}
	}

	for service, colocations := range v.CoLocatedServices {
		v.logger.Println(service, colocations)
	}
}

func (v *CoLocatedServiceInfosVisitor) VisitProcessNode(_ Visitor, n *ProcessNode) {
	v.curProcName = n.Name
	v.procServices[v.curProcName] = []string{}
	v.DefaultVisitor.VisitProcessNode(v, n)
}

func (v *CoLocatedServiceInfosVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.procServices[v.curProcName] = append(v.procServices[v.curProcName], n.Name)
}
