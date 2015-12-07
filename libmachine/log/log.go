package log

import "io"

var logger MachineLogger

func init() {
	logger = NewMachineLogger()
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func RedirectStdOutToStdErr() {
	logger.RedirectStdOutToStdErr()
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugf(fmtString string, args ...interface{}) {
	logger.Debugf(fmtString, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(fmtString string, args ...interface{}) {
	logger.Errorf(fmtString, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(fmtString string, args ...interface{}) {
	logger.Infof(fmtString, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(fmtString string, args ...interface{}) {
	logger.Fatalf(fmtString, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(fmtString string, args ...interface{}) {
	logger.Warnf(fmtString, args...)
}

func Logger() interface{} {
	return logger
}

func SetDebug(debug bool) {
	logger.SetDebug(debug)
}

func SetOutput(out io.Writer) {
	logger.SetOutput(out)
}
