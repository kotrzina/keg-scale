package wa

import (
	"github.com/sirupsen/logrus"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type Logger struct {
	l *logrus.Logger
}

func createLogger(l *logrus.Logger) *Logger {
	return &Logger{l: l}
}

func (l Logger) Warnf(msg string, args ...interface{}) {
	l.l.Warnf(msg, args...)
}

func (l Logger) Errorf(msg string, args ...interface{}) {
	l.l.Errorf(msg, args...)
}

func (l Logger) Fatalf(msg string, args ...interface{}) {
	l.l.Fatalf(msg, args...)
}

func (l Logger) Infof(msg string, args ...interface{}) {
	l.l.Infof(msg, args...)
}

func (l Logger) Debugf(msg string, args ...interface{}) {
	l.l.Debugf(msg, args...)
}

func (l Logger) Sub(_ string) waLog.Logger {
	l.l.Debug("Sub logger not implemented")
	return l
}
