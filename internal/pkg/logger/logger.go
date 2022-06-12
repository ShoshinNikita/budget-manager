package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

type Fields logrus.Fields

func New() Logger {
	log := logrus.New()
	log.SetReportCaller(false)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(formatter{})

	return logger{log}
}

// logger is a wrapper for logrus.FieldLogger that implements Logger interface
type logger struct {
	logrus.FieldLogger
}

func (l logger) WithField(key string, value interface{}) Logger {
	return logger{l.FieldLogger.WithField(key, value)}
}

func (l logger) WithFields(fields Fields) Logger {
	return logger{l.FieldLogger.WithFields(logrus.Fields(fields))}
}

func (l logger) WithError(err error) Logger {
	return logger{l.FieldLogger.WithError(err)}
}
