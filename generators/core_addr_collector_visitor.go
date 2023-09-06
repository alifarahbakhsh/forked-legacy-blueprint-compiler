package generators

import (
	"log"
)

type ConnInfo struct {
	Address  string
	Hostname string
	Port     int
}

type AddrCollectorVisitor struct {
	DefaultVisitor
	logger *log.Logger
	Addrs  map[string]ConnInfo
}

func NewAddrCollectorVisitor(logger *log.Logger) *AddrCollectorVisitor {
	return &AddrCollectorVisitor{logger: logger, Addrs: make(map[string]ConnInfo)}
}

func (v *AddrCollectorVisitor) VisitMillenialNode(_ Visitor, n *MillenialNode) {
	v.logger.Println("Starting AddrCollectorVisitor visit")
	v.DefaultVisitor.VisitMillenialNode(v, n)
	v.logger.Println("Ending AddrCollectorVisitor visit")
}

func (v *AddrCollectorVisitor) VisitFuncServiceNode(_ Visitor, n *FuncServiceNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitJaegerNode(_ Visitor, n *JaegerNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitZipkinNode(_ Visitor, n *ZipkinNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitXTraceNode(_ Visitor, n *XTraceNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitMemcachedNode(_ Visitor, n *MemcachedNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitRedisNode(_ Visitor, n *RedisNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitMongoDBNode(_ Visitor, n *MongoDBNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitRabbitMQNode(_ Visitor, n *RabbitMQNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitMySqlDBNode(_ Visitor, n *MySqlDBNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}

func (v *AddrCollectorVisitor) VisitConsulNode(_ Visitor, n *ConsulNode) {
	cinfo := ConnInfo{Address: n.DepInfo.Address, Port: n.DepInfo.Port, Hostname: n.DepInfo.Hostname}
	v.Addrs[n.Name] = cinfo
}
