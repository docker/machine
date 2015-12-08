package log

import "io"

var Logger MachineLogger

func init() {
	Logger = NewFmtMachineLogger()
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func RedirectStdOutToStdErr() {
	Logger.RedirectStdOutToStdErr()
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(fmtString string, args ...interface{}) {
	Logger.Debugf(fmtString, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(fmtString string, args ...interface{}) {
	Logger.Errorf(fmtString, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(fmtString string, args ...interface{}) {
	Logger.Infof(fmtString, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(fmtString string, args ...interface{}) {
	Logger.Fatalf(fmtString, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(fmtString string, args ...interface{}) {
	Logger.Warnf(fmtString, args...)
}

func SetDebug(debug bool) {
	Logger.SetDebug(debug)
}

func SetOutput(out io.Writer) {
	Logger.SetOutput(out)
}

func History() []string {
	return Logger.History()
}
