package log

import (
	"fmt"
	glog "log"
	"os"
)

const (
	// Prefix for the default (plugin) logger
	Prefix = "plugin-"
)

// Default logger implementation in case no logger is provided explicitly.
// This will simply use golang log on stdout
type defaultLogger struct {
	glogger *glog.Logger
}

func newDefaultLogger() Logger {
	return &defaultLogger{glogger: glog.New(os.Stdout, Prefix, 0)}
}

func (t *defaultLogger) log(level string, args ...interface{}) {
	t.glogger.Printf("%s: %s", level, fmt.Sprint(args...))
}

func (t *defaultLogger) logf(level string, fmtString string, args ...interface{}) {
	t.glogger.Printf("%s: %s", level, fmt.Sprintf(fmtString, args...))
}

func (t *defaultLogger) Debug(args ...interface{}) {
	t.log("Debug", args...)
}

func (t *defaultLogger) Debugf(fmtString string, args ...interface{}) {
	t.logf("Debug", fmtString, args...)
}

func (t *defaultLogger) Error(args ...interface{}) {
	t.log("Error", args...)
}

func (t *defaultLogger) Errorf(fmtString string, args ...interface{}) {
	t.logf("Error", fmtString, args...)
}

func (t *defaultLogger) Info(args ...interface{}) {
	t.log("Info", args...)
}

func (t *defaultLogger) Infof(fmtString string, args ...interface{}) {
	t.logf("Info", fmtString, args...)
}

func (t *defaultLogger) Fatal(args ...interface{}) {
	t.glogger.Fatalf("Fatal: %s", fmt.Sprint(args...))
}

func (t *defaultLogger) Fatalf(fmtString string, args ...interface{}) {
	t.glogger.Fatalf("Fatal: %s", fmt.Sprintf(fmtString, args...))
}

func (t *defaultLogger) Print(args ...interface{}) {
	t.log("Print", args...)
}

func (t *defaultLogger) Printf(fmtString string, args ...interface{}) {
	t.logf("Print", fmtString, args...)
}

func (t *defaultLogger) Warn(args ...interface{}) {
	t.log("Warn", args...)
}

func (t *defaultLogger) Warnf(fmtString string, args ...interface{}) {
	t.logf("Warn", fmtString, args...)
}

func (t *defaultLogger) Panic(args ...interface{}) {
	t.glogger.Panicf("Panic: %s", fmt.Sprint(args...))
}

func (t *defaultLogger) Panicf(fmtString string, args ...interface{}) {
	t.glogger.Panicf("Panic: %s", fmt.Sprintf(fmtString, args...))
}
