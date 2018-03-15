package registry

import (
	"fmt"
	"net"

	consul "github.com/hashicorp/consul/api"
)

// Client provides an interface for communicating with registry
type Client interface {
	Register(string, int) error
	Deregister(string) error
	Service(string) ([]string, error)
}

// NewClient returns an implementation of the Client interface, wrapping a
// concrete Consul client.
func NewClient(addr string) (Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &client{consul: c}, nil
}

type client struct {
	consul *consul.Client
}

// Register a service with registry
func (c *client) Register(name string, port int) error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Errorf("unable to determine local addr: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	reg := &consul.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Port:    port,
		Address: localAddr.IP.String(),
	}

	return c.consul.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (c *client) Deregister(id string) error {
	return c.consul.Agent().ServiceDeregister(id)
}

// Service returns the addresses for a given service
func (c *client) Service(name string) ([]string, error) {
	resp, _, err := c.consul.Health().Service(name, "", false, nil)
	if err != nil {
		return nil, err
	}

	ss := []string{}
	for _, s := range resp {
		ss = append(
			ss,
			fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port),
		)
	}
	return ss, nil
}
