package log

import (
	"io"

	"fmt"

	"sync"

	"github.com/Sirupsen/logrus"
)

type LogrusMachineLogger struct {
	history     []string
	historyLock sync.Locker
	logger      *logrus.Logger
}

// NewLogrusMachineLogger creates the MachineLogger implementation used by the docker-machine
func NewLogrusMachineLogger() MachineLogger {
	logrusLogger := logrus.New()
	logrusLogger.Level = logrus.InfoLevel
	logrusLogger.Formatter = new(MachineFormatter)
	return &LogrusMachineLogger{[]string{}, &sync.Mutex{}, logrusLogger}
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func (ml *LogrusMachineLogger) RedirectStdOutToStdErr() {
	ml.logger.Level = logrus.ErrorLevel
}

func (ml *LogrusMachineLogger) SetDebug(debug bool) {
	if debug {
		ml.logger.Level = logrus.DebugLevel
	} else {
		ml.logger.Level = logrus.InfoLevel
	}
}

func (ml *LogrusMachineLogger) SetOutput(out io.Writer) {
	ml.logger.Out = out
}

func (ml *LogrusMachineLogger) Logger() *logrus.Logger {
	return ml.logger
}

func (ml *LogrusMachineLogger) Debug(args ...interface{}) {
	ml.record(args...)
	ml.logger.Debug(args...)
}

func (ml *LogrusMachineLogger) Debugf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.logger.Debugf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Error(args ...interface{}) {
	ml.record(args...)
	ml.logger.Error(args...)
}

func (ml *LogrusMachineLogger) Errorf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.logger.Errorf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Info(args ...interface{}) {
	ml.record(args...)
	ml.logger.Info(args...)
}

func (ml *LogrusMachineLogger) Infof(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.logger.Infof(fmtString, args...)
}

func (ml *LogrusMachineLogger) Fatal(args ...interface{}) {
	ml.record(args...)
	ml.logger.Fatal(args...)
}

func (ml *LogrusMachineLogger) Fatalf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.logger.Fatalf(fmtString, args...)
}

func (ml *LogrusMachineLogger) Warn(args ...interface{}) {
	ml.record(args...)
	ml.logger.Warn(args...)
}

func (ml *LogrusMachineLogger) Warnf(fmtString string, args ...interface{}) {
	ml.recordf(fmtString, args...)
	ml.logger.Warnf(fmtString, args...)
}

func (ml *LogrusMachineLogger) History() []string {
	return ml.history
}

func (ml *LogrusMachineLogger) record(args ...interface{}) {
	ml.historyLock.Lock()
	defer ml.historyLock.Unlock()
	ml.history = append(ml.history, fmt.Sprint(args...))
}

func (ml *LogrusMachineLogger) recordf(fmtString string, args ...interface{}) {
	ml.historyLock.Lock()
	defer ml.historyLock.Unlock()
	ml.history = append(ml.history, fmt.Sprintf(fmtString, args...))
}
