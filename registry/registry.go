package registry

import (
	"fmt"
	"net"

	consul "github.com/hashicorp/consul/api"
	"github.com/segmentio/ksuid"
)

// NewClient returns a new Client with connection to consul
func NewClient(addr string) (*Client, error) {
	cfg := consul.DefaultConfig()
	cfg.Address = addr

	c, err := consul.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

// Client provides an interface for communicating with registry
type Client struct {
	*consul.Client
}

// Register a service with registry
func (c *Client) Register(name string, port int) (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("unable to determine local addr: %v", err)
	}
	defer conn.Close()

	var (
		uuid      = fmt.Sprintf("%s-%s", name, ksuid.New().String())
		localAddr = conn.LocalAddr().(*net.UDPAddr)
	)

	reg := &consul.AgentServiceRegistration{
		ID:      uuid,
		Name:    name,
		Port:    port,
		Address: localAddr.IP.String(),
	}

	return uuid, c.Agent().ServiceRegister(reg)
}

// Deregister removes the service address from registry
func (c *Client) Deregister(id string) error {
	return c.Agent().ServiceDeregister(id)
}
