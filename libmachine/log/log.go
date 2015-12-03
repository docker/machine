package log

import (
	"os"
	"sync"
)

var (
	l = StandardLogger{
		OutWriter: os.Stdout,
		ErrWriter: os.Stderr,
		mu:        &sync.Mutex{},
	}
	IsDebug = false
)

// RedirectStdOutToStdErr prevents any log from corrupting the output
func RedirectStdOutToStdErr() {
	l.OutWriter = l.ErrWriter
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

func Info(args ...interface{}) {
	l.Info(args...)
}

func Infof(fmtString string, args ...interface{}) {
	l.Infof(fmtString, args...)
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

func GetStandardLogger() *StandardLogger {
	return &l
}
