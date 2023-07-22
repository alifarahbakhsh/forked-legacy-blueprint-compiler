package generators

import (
	"log"

	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/generators/deploy"
	"gitlab.mpi-sws.org/cld/blueprint/blueprint-compiler/parser"
)

type BasicDeployVisitor struct {
	DefaultVisitor
	logger        *log.Logger
	config        *parser.Config
	addresses     map[string]parser.Address
	envVars       map[string][]parser.EnvVariable
	portAuthority *deploy.PortAuthority
}

func NewBasicDeployVisitor(logger *log.Logger, config *parser.Config, portAuthority *deploy.PortAuthority) *BasicDeployVisitor {
	addresses := make(map[string]parser.Address)
	for _, addr := range config.Addresses {
		addresses[addr.Name] = addr
	}
	envVars := make(map[string][]parser.EnvVariable)
	for _, envr := range config.Environment {
		envVars[envr.Name] = envr.Variables
	}
	return &BasicDeployVisitor{logger: logger, config: config, addresses: addresses, portAuthority: portAuthority, envVars: envVars}
}

func (v *BasicDeployVisitor) modifyEnvMap(nodeName string, envMap map[string]string) {
	if vars, ok := v.envVars[nodeName]; ok {
		for _, envvar := range vars {
			envMap[envvar.Name] = envvar.Value
		}
	}
}

func (v *BasicDeployVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting BasicDeployVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending BasicDeployVisitor visit")
}

func (v *BasicDeployVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 8000
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitQueueServiceNode(_ Visitor, n *QueueServiceNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 8000
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitJaegerNode(_ Visitor, n *JaegerNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 14268
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitZipkinNode(_ Visitor, n *ZipkinNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 9411
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitXTraceNode(_ Visitor, n *XTraceNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 5563
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitMemcachedNode(_ Visitor, n *MemcachedNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 11211
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitRedisNode(_ Visitor, n *RedisNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 6379
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitMongoDBNode(_ Visitor, n *MongoDBNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		defaultPort := 27017
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitRabbitMQNode(_ Visitor, n *RabbitMQNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		n.DepInfo.Hostname = defaultAddress
		defaultPort := 5672
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}

func (v *BasicDeployVisitor) VisitMySqlDBNode(_ Visitor, n *MySqlDBNode) {
	v.modifyEnvMap(n.Name, n.DepInfo.EnvVars)
	if addr, ok := v.addresses[n.Name]; ok {
		n.DepInfo.Address = addr.Address
		n.DepInfo.Hostname = addr.Hostname
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(addr.Address, addr.Port)
	} else {
		defaultAddress := "localhost"
		n.DepInfo.Hostname = defaultAddress
		defaultPort := 3306
		n.DepInfo.Address = defaultAddress
		n.DepInfo.Hostname = defaultAddress
		n.DepInfo.Port = v.portAuthority.GetAvailablePort(defaultAddress, defaultPort)
	}
	v.logger.Println("Assigned address:", n.DepInfo.Address, ":", n.DepInfo.Port, "to", n.Name)
}
