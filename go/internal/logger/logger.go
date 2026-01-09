package logger

import (
	"log/slog"
	"os"

	"go.temporal.io/sdk/log"
)

var Log *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
}

type temporalLogger struct{}

func NewTemporalLogger() log.Logger {
	return &temporalLogger{}
}

func (l *temporalLogger) Debug(msg string, keyvals ...interface{}) {}
func (l *temporalLogger) Info(msg string, keyvals ...interface{})  {}
func (l *temporalLogger) Warn(msg string, keyvals ...interface{}) {
	Log.Warn(msg, keyvals...)
}

func (l *temporalLogger) Error(msg string, keyvals ...interface{}) {
	Log.Error(msg, keyvals...)
}
