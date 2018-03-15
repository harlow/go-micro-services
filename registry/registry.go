package registry

import (
	"fmt"
	"net"

	"github.com/hashicorp/consul/api"
)

// NewClient returns a new Client with connection to consul
func NewClient(addr string) (*Client, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

// Client provides an interface for communicating with registry
type Client struct {
	*api.Client
}

// Register a service with registry
func (c *Client) Register(name string, port int) error {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return fmt.Errorf("unable to determine local addr: %v", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	reg := &api.AgentServiceRegistration{
		ID:      name,
		Name:    name,
		Port:    port,
		Address: localAddr.IP.String(),
	}

	return c.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (c *Client) Deregister(id string) error {
	return c.Agent().ServiceDeregister(id)
}
