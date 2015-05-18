package log

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"
)

const (
	debugEnvKey = "DEBUG"
)

var (
	l    = NewStandardLogger(os.Stdout, os.Stderr)
	lock = &sync.Mutex{}
)

// Why the interface?  We may only want to print to STDOUT and STDERR for now,
// but it won't neccessarily be that way forever.  This interface is intended
// to provide a "framework" for a variety of different logging types in the
// future (log to file, log to logstash, etc.) There could be a driver model
// similar to what is done with OS or machine providers.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Debugln(...interface{})

	Error(...interface{})
	Errorf(string, ...interface{})
	Errorln(...interface{})

	Info(...interface{})
	Infof(string, ...interface{})
	Infoln(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Warn(...interface{})
	Warnf(string, ...interface{})
	Warnln(...interface{})

	// WithFields returns a new Logger with the specified field.
	// The original logger will not be affected.
	WithField(fieldName string, field interface{}) Logger
	// WithFields returns a new Logger with the specified fields.
	// The original logger will not be affected.
	WithFields(Fields) Logger
}

// NewStandardLogger returns a new standard logger for logging.
// stderr is allowed to be nil, in this case, all logging will go to stdout.
func NewStandardLogger(stdout io.Writer, stderr io.Writer) Logger {
	return newStandardLogger(stdout, stderr, make(Fields))
}

// SetLogger sets the logger used by docker-machine.
func SetLogger(logger Logger) {
	lock.Lock()
	defer lock.Unlock()
	l = logger
}

type Fields map[string]interface{}

func Debug(args ...interface{}) {
	l.Debug(args...)
}

func Debugf(fmtString string, args ...interface{}) {
	l.Debugf(fmtString, args...)
}

func Debugln(args ...interface{}) {
	l.Debugln(args...)
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

func Fatalln(args ...interface{}) {
	l.Fatalln(args...)
}

func Print(args ...interface{}) {
	l.Print(args...)
}

func Printf(fmtString string, args ...interface{}) {
	l.Printf(fmtString, args...)
}

func Println(args ...interface{}) {
	l.Println(args...)
}

func Warn(args ...interface{}) {
	l.Warn(args...)
}

func Warnf(fmtString string, args ...interface{}) {
	l.Warnf(fmtString, args...)
}

func Warnln(args ...interface{}) {
	l.Warnln(args...)
}

func WithField(fieldName string, field interface{}) Logger {
	return l.WithField(fieldName, field)
}

func WithFields(fields Fields) Logger {
	return l.WithFields(fields)
}

// TODO: I think this is superflous and can be replaced by one check for if
// debug is on that sets a variable in this module.
func isDebug() bool {
	debugEnv := os.Getenv(debugEnvKey)
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

func copyFields(fields Fields) Fields {
	c := make(Fields)
	for k, v := range fields {
		c[k] = v
	}
	return c
}
