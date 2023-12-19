package logger

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger of module
type Logger interface {
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	Fatalln(args ...interface{})
}

type logger struct {
	entry *logrus.Entry
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{l.entry.WithField(key, value)}
}

func (l *logger) WithError(err error) Logger {
	return &logger{l.entry.WithError(err)}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logger) Debugln(args ...interface{}) {
	l.entry.Debugln(args...)
}

func (l *logger) Infoln(args ...interface{}) {
	l.entry.Infoln(args...)
}

func (l *logger) Warnln(args ...interface{}) {
	l.entry.Warnln(args...)
}

func (l *logger) Errorln(args ...interface{}) {
	l.entry.Errorln(args...)
}

func (l *logger) Fatalln(args ...interface{}) {
	l.entry.Fatalln(args...)
}

func newFormatter(format string) logrus.Formatter {
	var formatter logrus.Formatter
	if strings.ToLower(format) == "json" {
		formatter = &logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano}
	} else {
		formatter = &logrus.TextFormatter{TimestampFormat: time.RFC3339Nano, FullTimestamp: true, DisableColors: true}
	}
	return formatter
}

// New create a new logger
func New(c Config, fields ...string) (Logger, error) {
	persistenceType := persistenceFile

	if c.Store == persistenceFile ||
		c.Store == persistenceMongo ||
		c.Store == persistencePG {
		persistenceType = c.Store
	}

	return getPersistence(persistenceType, c, fields...)
}
