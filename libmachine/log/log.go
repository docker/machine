package log

import (
	"io"
	"os"
	"sync"
)

var (
	l = StandardLogger{
		mu: &sync.Mutex{},
	}
	IsDebug = false
)

type Fields map[string]interface{}

func init() {
	// TODO: Is this really the best approach?  I worry that it will create
	// implicit behavior which may be problmatic for users of the lib.
	SetOutWriter(os.Stdout)
	SetErrWriter(os.Stderr)
}

func SetOutWriter(w io.Writer) {
	l.OutWriter = w
}

func SetErrWriter(w io.Writer) {
	l.ErrWriter = w
}

func Debug(args ...interface{}) {
	l.Debug(args...)
}

func Debugf(fmtString string, args ...interface{}) {
	l.Debugf(fmtString, args...)
}

func Error(args ...interface{}) {
	l.Error(args...)
}

func Errorf(fmtString string, args ...interface{}) {
	l.Errorf(fmtString, args...)
}

func Errorln(args ...interface{}) {
	l.Errorln(args...)
}

func Info(args ...interface{}) {
	l.Info(args...)
}

func Infof(fmtString string, args ...interface{}) {
	l.Infof(fmtString, args...)
}

func Infoln(args ...interface{}) {
	l.Infoln(args...)
}

func Fatal(args ...interface{}) {
	l.Fatal(args...)
}

func Fatalf(fmtString string, args ...interface{}) {
	l.Fatalf(fmtString, args...)
}

func Warn(args ...interface{}) {
	l.Warn(args...)
}

func Warnf(fmtString string, args ...interface{}) {
	l.Warnf(fmtString, args...)
}

func WithField(fieldName string, field interface{}) Logger {
	return l.WithFields(Fields{
		fieldName: field,
	})
}

func WithFields(fields Fields) Logger {
	return l.WithFields(fields)
}
