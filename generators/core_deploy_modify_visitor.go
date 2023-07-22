package generators

import (
	"log"
)

type DeployModifierVisitor struct {
	DefaultVisitor
	logger *log.Logger
	modregistry *ModifierRegistry
}

func NewDeployModifierVisitor(logger *log.Logger, modregistry *ModifierRegistry) *DeployModifierVisitor {
	return &DeployModifierVisitor{DefaultVisitor{}, logger, modregistry}
}

func (v *DeployModifierVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting DeployModifierVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending DeployModifierVisitor visit")
}

func (v *DeployModifierVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.logger.Println("Applying Deploymend Modifiers for", n.Name)
	for _, modifier := range v.modregistry.OrderModifiers(n.ServerModifiers) {
		if !modifier.IsDeployerModifier() {
			// Only need to apply the deployer modifier in this visit
			continue
		}
		v.logger.Println("Applying Modifier:", modifier.GetName())
		modifier.ModifyDeployInfo(n.DepInfo)
	}
}