package log

var logger Logger

// SetLogger sets the logger for libmachine, must implement the Logger interface
func SetLogger(lgr Logger) {
	logger = lgr
}

// GetLogger gets the currently set logger for libmachine
func GetLogger() Logger {
	if logger == nil {
		logger = newDefaultLogger()
	}
	return logger
}

// Debug logs a message at level Debug
func Debug(args ...interface{}) {
	GetLogger().Debug(args...)
}

// Debugf logs a formatted message at level Debug
func Debugf(fmtString string, args ...interface{}) {
	GetLogger().Debugf(fmtString, args...)
}

// Error logs a message at level Error
func Error(args ...interface{}) {
	GetLogger().Error(args...)
}

// Errorf logs a formatted message at level Error
func Errorf(fmtString string, args ...interface{}) {
	GetLogger().Errorf(fmtString, args...)
}

// Info logs a message at level Info
func Info(args ...interface{}) {
	GetLogger().Info(args...)
}

// Infof logs a formatted message at level Info
func Infof(fmtString string, args ...interface{}) {
	GetLogger().Infof(fmtString, args...)
}

// Fatal logs a message at level Fatal
func Fatal(args ...interface{}) {
	GetLogger().Fatal(args...)
}

// Fatalf logs a formatted message at level Fatal
func Fatalf(fmtString string, args ...interface{}) {
	GetLogger().Fatalf(fmtString, args...)
}

// Print logs a message at level Info
func Print(args ...interface{}) {
	GetLogger().Print(args...)
}

// Printf logs a formatted message at level Info
func Printf(fmtString string, args ...interface{}) {
	GetLogger().Printf(fmtString, args...)
}

// Warn logs a message at level Warn
func Warn(args ...interface{}) {
	GetLogger().Warn(args...)
}

// Warnf logs a formatted message at level Warn
func Warnf(fmtString string, args ...interface{}) {
	GetLogger().Warnf(fmtString, args...)
}
