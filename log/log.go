package log

import (
	"github.com/Sirupsen/logrus"
	libmachinelog "github.com/docker/machine/libmachine/log"
	"io/ioutil"
	"os"
)

var logger *logrus.Logger
var stdlog *logrus.Logger
var errlog *logrus.Logger

// This is our main hook to funnel entries to the stderr and stdout loggers
type logrusHook struct {
}

func (h *logrusHook) Fire(e *logrus.Entry) error {
	e.Logger = stdlog
	switch e.Level {
	case logrus.DebugLevel:
		e.Logger = errlog
	case logrus.FatalLevel:
		e.Logger = errlog
	default:
		e.Logger = stdlog
	}
	return nil
}

func (h *logrusHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
}

func init() {
	// Define the main logger that:
	// - outputs to dev null
	// - defaults to Info
	// - attach our hook
	logger = logrus.New()
	logger.Out = ioutil.Discard
	logger.Level = logrus.DebugLevel
	logger.Hooks.Add(&logrusHook{})

	// Declare it as the main logger for libmachine
	libmachinelog.SetLogger(logger)

	// Create and configure the stderr and stdout loggers, with machine specific formatters
	errlog = logrus.New()
	errlog.Out = os.Stderr
	errlog.Formatter = new(machineFormatter)
	errlog.Level = logrus.InfoLevel

	stdlog = logrus.New()
	stdlog.Out = os.Stdout
	stdlog.Formatter = new(machineFormatter)
	stdlog.Level = logrus.InfoLevel
}

// GetMain returns the main logger, useful to set logging level overall
func GetMain() *logrus.Logger {
	return logger
}

// GetStd returns the standard output logger
func GetStd() *logrus.Logger {
	return stdlog
}

// GetErr returns the error output logger
func GetErr() *logrus.Logger {
	return errlog
}
