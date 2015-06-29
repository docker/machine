package log

import (
	"fmt"
	"os"
	"strconv"
)

// Why the interface?  We may only want to print to STDOUT and STDERR for now,
// but it won't neccessarily be that way forever.  This interface is intended
// to provide a "framework" for a variety of different logging types in the
// future (log to file, log to logstash, etc.) There could be a driver model
// similar to what is done with OS or machine providers.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})

	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})

	Print(...interface{})
	Printf(string, ...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})

	WithFields(Fields) Logger
}

var (
	l = TerminalLogger{}
)

// TODO: I think this is superflous and can be replaced by one check for if
// debug is on that sets a variable in this module.
func isDebug() bool {
	debugEnv := os.Getenv("DEBUG")
	if debugEnv != "" {
		showDebug, err := strconv.ParseBool(debugEnv)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing boolean value from DEBUG: %s", err)
			os.Exit(1)
		}
		return showDebug
	}
	return false
}

type Fields map[string]interface{}

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

func Print(args ...interface{}) {
	l.Print(args...)
}

func Printf(fmtString string, args ...interface{}) {
	l.Printf(fmtString, args...)
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
