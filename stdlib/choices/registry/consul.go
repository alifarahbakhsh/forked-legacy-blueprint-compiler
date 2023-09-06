package registry

import (
	"strconv"

	consul "github.com/hashicorp/consul/api"
)

type ConsulRegistry struct {
	client *consul.Client
}

func NewConsulRegistry(address string, port int) (*ConsulRegistry, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = address + ":" + strconv.Itoa(port)

	client, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &ConsulRegistry{client: client}, nil
}

func (r *ConsulRegistry) Register(ID string, name string, address string, port int64) error {
	reg := &consul.AgentServiceRegistration{
		ID:      ID,
		Name:    name,
		Port:    int(port),
		Address: address,
	}
	return r.client.Agent().ServiceRegister(reg)
}
