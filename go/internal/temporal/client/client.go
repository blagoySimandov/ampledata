package client

import (
	"fmt"

	"go.temporal.io/sdk/client"
)

// NewClient creates a new Temporal client with the given configuration
func NewClient(hostPort, namespace string) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}
	return c, nil
}
