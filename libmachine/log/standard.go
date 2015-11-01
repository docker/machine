package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
)

type StandardLogger struct {
	// fieldOut is used to do log.WithFields correctly
	fieldOut  string
	OutWriter io.Writer
	ErrWriter io.Writer
	mu        *sync.Mutex
}

func (t StandardLogger) log(args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprint(t.OutWriter, args...)
	fmt.Fprint(t.OutWriter, t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t StandardLogger) logf(fmtString string, args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprintf(t.OutWriter, fmtString, args...)
	fmt.Fprint(t.OutWriter, "\n")
	t.fieldOut = ""
}

func (t StandardLogger) err(args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprint(t.ErrWriter, args...)
	fmt.Fprint(t.ErrWriter, t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t StandardLogger) errf(fmtString string, args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprintf(t.ErrWriter, fmtString, args...)
	fmt.Fprint(t.ErrWriter, t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t StandardLogger) Debug(args ...interface{}) {
	if IsDebug {
		t.err(args...)
	}
}

func (t StandardLogger) Debugf(fmtString string, args ...interface{}) {
	if IsDebug {
		t.errf(fmtString, args...)
	}
}

func (t StandardLogger) Error(args ...interface{}) {
	t.err(args...)
}

func (t StandardLogger) Errorf(fmtString string, args ...interface{}) {
	t.errf(fmtString, args...)
}

func (t StandardLogger) Errorln(args ...interface{}) {
	t.err(args...)
}

func (t StandardLogger) Info(args ...interface{}) {
	t.log(args...)
}

func (t StandardLogger) Infof(fmtString string, args ...interface{}) {
	t.logf(fmtString, args...)
}

func (t StandardLogger) Infoln(args ...interface{}) {
	t.log(args...)
}

func (t StandardLogger) Fatal(args ...interface{}) {
	t.err(args...)
	os.Exit(1)
}

func (t StandardLogger) Fatalf(fmtString string, args ...interface{}) {
	t.errf(fmtString, args...)
	os.Exit(1)
}

func (t StandardLogger) Print(args ...interface{}) {
	t.log(args...)
}

func (t StandardLogger) Printf(fmtString string, args ...interface{}) {
	t.logf(fmtString, args...)
}

func (t StandardLogger) Warn(args ...interface{}) {
	fmt.Print("WARNING >>> ")
	t.log(args...)
}

func (t StandardLogger) Warnf(fmtString string, args ...interface{}) {
	fmt.Print("WARNING >>> ")
	t.logf(fmtString, args...)
}

func (t StandardLogger) WithFields(fields Fields) Logger {
	// When the user calls WithFields, we make a string which gets appended
	// to the output of the final [Info|Warn|Error] call for the
	// descriptive fields.  Because WithFields returns the proper Logger
	// (with the fieldOut string set correctly), the logrus syntax will
	// still work.
	kvpairs := []string{}

	// Why the string slice song and dance?  Because Go's map iteration
	// order is random, we will get inconsistent results if we don't sort
	// the fields (or their resulting string K/V pairs, like we have here).
	// Otherwise, we couldn't test this reliably.
	for k, v := range fields {
		kvpairs = append(kvpairs, fmt.Sprintf("%s=%v", k, v))
	}

	sort.Strings(kvpairs)

	// TODO:
	// 1. Is this thread-safe?
	// 2. Add more tabs?
	t.fieldOut = "\t\t"

	for _, s := range kvpairs {
		// TODO: Is %v the correct format string here?
		t.fieldOut = fmt.Sprintf("%s %s", t.fieldOut, s)
	}

	return t
}
