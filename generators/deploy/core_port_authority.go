package deploy

import (
	"log"
)

const (
	MAX_PORT_NUMBER = 65535
	MIN_PORT_NUMBER = 1024 // 1-1023 require root access so we won't use them
)

type PortAuthority struct {
	logger *log.Logger
	PortsInUse map[string]map[int]bool
}

func NewPortAuthority(logger *log.Logger) *PortAuthority {
	return &PortAuthority{logger: logger, PortsInUse: make(map[string]map[int]bool)}
}

func (pa *PortAuthority) isPortAvailable(hostname string, port int) bool {
	if v, ok := pa.PortsInUse[hostname]; ok {
		if _, ok2 := v[port]; ok2 {
			return false
		} else {
			v[port] = true
			return true
		}
	} else {
		pa.PortsInUse[hostname] = make(map[int]bool)
		pa.PortsInUse[hostname][port] = true
		return true
	}
	return true
}

func (pa *PortAuthority) GetAvailablePort(hostname string, port int) int {
	if pa.isPortAvailable(hostname, port) {
		return port
	}

	p := port + 1
	checked_ports := make(map[int]bool)
	checked_ports[port] = true
	for p >= MIN_PORT_NUMBER && p <= MAX_PORT_NUMBER {
		if p == port {
			// Well, we have circled all the way
			pa.logger.Fatal("Hostname:", hostname, "has no free port available for assignment")
		}
		if pa.isPortAvailable(hostname, p) {
			break
		}
		checked_ports[p] = true
		if p == MAX_PORT_NUMBER {
			p = MIN_PORT_NUMBER
		} else {
			p += 1
		}
	}
	return p
}