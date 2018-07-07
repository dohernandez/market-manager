package logger

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

type (
	// IndexContext is a key type for context values
	IndexContext int
	// Logger defines a logging interface
	Logger logrus.FieldLogger
)

// Key constants for context values
const (
	TraceIDKey IndexContext = iota
)

var (
	baseLogger logrus.FieldLogger
)

func init() {
	baseLogger = logrus.New()
}

// SetLogger receives a logger and stores it as the base logger for the package.
func SetLogger(logger logrus.FieldLogger) {
	baseLogger = logger
}

// FromContext returns an logger with all values from context loaded on it.
func FromContext(ctx context.Context) logrus.FieldLogger {
	log := baseLogger
	if id, err := traceID(ctx); err == nil {
		log = log.WithField("traceID", id)
	}

	return log
}

func traceID(ctx context.Context) (string, error) {
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok {
		return "", errors.New("No TraceID in the context")
	}

	return traceID, nil
}
