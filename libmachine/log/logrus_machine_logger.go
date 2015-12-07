package log

import (
	"io"

	"github.com/Sirupsen/logrus"
)

type LogrusMachineLogger struct {
	logger *logrus.Logger
}

func NewMachineLogger() MachineLogger {
	logrusLogger := logrus.New()
	logrusLogger.Level = logrus.InfoLevel
	logrusLogger.Formatter = new(MachineFormatter)
	return LogrusMachineLogger{logrusLogger}
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func (ml LogrusMachineLogger) RedirectStdOutToStdErr() {
	ml.logger.Level = logrus.ErrorLevel
}

func (ml LogrusMachineLogger) SetDebug(debug bool) {
	if debug {
		ml.logger.Level = logrus.DebugLevel
	} else {
		ml.logger.Level = logrus.InfoLevel
	}
}

func (ml LogrusMachineLogger) SetOutput(out io.Writer) {
	ml.logger.Out = out
}

func (ml LogrusMachineLogger) Logger() interface{} {
	return ml.logger
}

func (ml LogrusMachineLogger) Debug(args ...interface{}) {
	ml.logger.Debug(args...)
}

func (ml LogrusMachineLogger) Debugf(fmtString string, args ...interface{}) {
	ml.logger.Debugf(fmtString, args...)
}

func (ml LogrusMachineLogger) Error(args ...interface{}) {
	ml.logger.Error(args...)
}

func (ml LogrusMachineLogger) Errorf(fmtString string, args ...interface{}) {
	ml.logger.Errorf(fmtString, args...)
}

func (ml LogrusMachineLogger) Info(args ...interface{}) {
	ml.logger.Info(args...)
}

func (ml LogrusMachineLogger) Infof(fmtString string, args ...interface{}) {
	ml.logger.Infof(fmtString, args...)
}

func (ml LogrusMachineLogger) Fatal(args ...interface{}) {
	ml.logger.Fatal(args...)
}

func (ml LogrusMachineLogger) Fatalf(fmtString string, args ...interface{}) {
	ml.logger.Fatalf(fmtString, args...)
}

func (ml LogrusMachineLogger) Warn(args ...interface{}) {
	ml.logger.Warn(args...)
}

func (ml LogrusMachineLogger) Warnf(fmtString string, args ...interface{}) {
	ml.logger.Warnf(fmtString, args...)
}
