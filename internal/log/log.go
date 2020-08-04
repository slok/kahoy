package log

import (
	"github.com/sirupsen/logrus"
)

// Kv is a helper type for structured logging fields usage.
type Kv = map[string]interface{}

// Logger is the interface that the loggers used by the library will use.
type Logger interface {
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	WithValues(values map[string]interface{}) Logger
}

// Noop logger doesn't log anything.
const Noop = noop(0)

type noop int

func (n noop) Infof(format string, args ...interface{})    {}
func (n noop) Warningf(format string, args ...interface{}) {}
func (n noop) Errorf(format string, args ...interface{})   {}
func (n noop) Debugf(format string, args ...interface{})   {}
func (n noop) WithValues(map[string]interface{}) Logger    { return n }

type logger struct {
	*logrus.Entry
}

// NewLogrus returns a new log.Logger for a logrus implementation.
func NewLogrus(l *logrus.Entry) Logger {
	return logger{Entry: l}
}

func (l logger) WithValues(kv map[string]interface{}) Logger {
	newLogger := l.Entry.WithFields(kv)
	return NewLogrus(newLogger)
}
