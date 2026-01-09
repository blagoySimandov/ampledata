package client

import (
	"fmt"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"go.temporal.io/sdk/client"
)

func NewClient(hostPort, namespace string) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		Logger:    logger.NewTemporalLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}
	return c, nil
}
