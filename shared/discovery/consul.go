package discovery

import (
	"fmt"
	"log"
	"os"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

// ServiceRegistry handles service registration and discovery with Consul
type ServiceRegistry struct {
	client *consulapi.Client
}

// ServiceConfig holds configuration for service registration
type ServiceConfig struct {
	ID              string
	Name            string
	Address         string
	Port            int
	Tags            []string
	HealthCheckPath string
	HealthCheckInterval time.Duration
	DeregisterAfter time.Duration
}

// NewServiceRegistry creates a new Consul service registry client
func NewServiceRegistry(consulAddr string) (*ServiceRegistry, error) {
	if consulAddr == "" {
		consulAddr = os.Getenv("CONSUL_ADDR")
		if consulAddr == "" {
			consulAddr = "localhost:8500"
		}
	}

	config := consulapi.DefaultConfig()
	config.Address = consulAddr

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}

	return &ServiceRegistry{client: client}, nil
}

// Register registers a service with Consul
func (sr *ServiceRegistry) Register(config ServiceConfig) error {
	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 10 * time.Second
	}
	if config.DeregisterAfter == 0 {
		config.DeregisterAfter = 1 * time.Minute
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      config.ID,
		Name:    config.Name,
		Address: config.Address,
		Port:    config.Port,
		Tags:    config.Tags,
		Check: &consulapi.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s:%d%s", config.Address, config.Port, config.HealthCheckPath),
			Interval:                       config.HealthCheckInterval.String(),
			Timeout:                        "5s",
			DeregisterCriticalServiceAfter: config.DeregisterAfter.String(),
		},
	}

	err := sr.client.Agent().ServiceRegister(registration)
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	log.Printf("Service %s registered with Consul (ID: %s)", config.Name, config.ID)
	return nil
}

// Deregister removes a service from Consul
func (sr *ServiceRegistry) Deregister(serviceID string) error {
	err := sr.client.Agent().ServiceDeregister(serviceID)
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	log.Printf("Service %s deregistered from Consul", serviceID)
	return nil
}

// DiscoverService finds a healthy instance of a service
func (sr *ServiceRegistry) DiscoverService(serviceName string) (string, error) {
	services, _, err := sr.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to discover service: %w", err)
	}

	if len(services) == 0 {
		return "", fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	// Return the first healthy service
	service := services[0]
	address := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)

	return address, nil
}

// DiscoverAllServices finds all healthy instances of a service
func (sr *ServiceRegistry) DiscoverAllServices(serviceName string) ([]string, error) {
	services, _, err := sr.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("no healthy instances of service %s found", serviceName)
	}

	addresses := make([]string, 0, len(services))
	for _, service := range services {
		address := fmt.Sprintf("http://%s:%d", service.Service.Address, service.Service.Port)
		addresses = append(addresses, address)
	}

	return addresses, nil
}

// GetKV retrieves a key-value pair from Consul
func (sr *ServiceRegistry) GetKV(key string) (string, error) {
	pair, _, err := sr.client.KV().Get(key, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if pair == nil {
		return "", fmt.Errorf("key %s not found", key)
	}

	return string(pair.Value), nil
}

// PutKV stores a key-value pair in Consul
func (sr *ServiceRegistry) PutKV(key, value string) error {
	pair := &consulapi.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := sr.client.KV().Put(pair, nil)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}

	return nil
}

// DeleteKV removes a key-value pair from Consul
func (sr *ServiceRegistry) DeleteKV(key string) error {
	_, err := sr.client.KV().Delete(key, nil)
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}
