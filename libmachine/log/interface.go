package log

// Logger defines the interface that a logger must implement to be used by libmachine consumers.
type Logger interface {
	Print(args ...interface{})
	Printf(format string, args ...interface{})

	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})

	Panic(args ...interface{})
	Panicf(format string, args ...interface{})

	Debug(args ...interface{})
	Debugf(format string, args ...interface{})

	Error(args ...interface{})
	Errorf(format string, args ...interface{})

	Info(args ...interface{})
	Infof(format string, args ...interface{})

	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
}
