package log

import "io"

type MachineLogger interface {
	RedirectStdOutToStdErr()

	SetDebug(debug bool)

	SetOutput(io.Writer)

	Debug(args ...interface{})
	Debugf(fmtString string, args ...interface{})

	Error(args ...interface{})
	Errorf(fmtString string, args ...interface{})

	Info(args ...interface{})
	Infof(fmtString string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(fmtString string, args ...interface{})

	Warn(args ...interface{})
	Warnf(fmtString string, args ...interface{})

	Logger() interface{}
}
