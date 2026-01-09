package client

import (
	"fmt"

	"github.com/blagoySimandov/ampledata/go/internal/logger"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
)

// temporalLogger adapts our logger to Temporal's logger interface
type temporalLogger struct{}

// NewTemporalLogger creates a new Temporal logger adapter
func NewTemporalLogger() log.Logger {
	return &temporalLogger{}
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *temporalLogger) Info(msg string, keyvals ...interface{})  {}
func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	logger.Log.Warn(msg, keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	logger.Log.Error(msg, keyvals...)
}

func NewClient(hostPort, namespace string) (client.Client, error) {
	c, err := client.Dial(client.Options{
		HostPort:  hostPort,
		Namespace: namespace,
		Logger:    NewTemporalLogger(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create Temporal client: %w", err)
	}
	return c, nil
}
