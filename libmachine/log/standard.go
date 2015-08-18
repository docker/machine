package log

import (
	"fmt"
	"os"
	"sort"
)

type TerminalLogger struct {
	// fieldOut is used to do log.WithFields correctly
	fieldOut string
}

func (t TerminalLogger) log(args ...interface{}) {
	fmt.Print(args...)
	fmt.Print(t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t TerminalLogger) logf(fmtString string, args ...interface{}) {
	fmt.Printf(fmtString, args...)
	fmt.Print(t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t TerminalLogger) err(args ...interface{}) {
	fmt.Fprint(os.Stderr, args...)
	fmt.Fprint(os.Stderr, t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t TerminalLogger) errf(fmtString string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtString, args...)
	fmt.Fprint(os.Stderr, t.fieldOut, "\n")
	t.fieldOut = ""
}

func (t TerminalLogger) Debug(args ...interface{}) {
	if IsDebug {
		t.log(args...)
	}
}

func (t TerminalLogger) Debugf(fmtString string, args ...interface{}) {
	if IsDebug {
		t.logf(fmtString, args...)
	}
}

func (t TerminalLogger) Error(args ...interface{}) {
	t.err(args...)
}

func (t TerminalLogger) Errorf(fmtString string, args ...interface{}) {
	t.errf(fmtString, args...)
}

func (t TerminalLogger) Info(args ...interface{}) {
	t.log(args...)
}

func (t TerminalLogger) Infof(fmtString string, args ...interface{}) {
	t.logf(fmtString, args...)
}

func (t TerminalLogger) Fatal(args ...interface{}) {
	t.err(args...)
	os.Exit(1)
}

func (t TerminalLogger) Fatalf(fmtString string, args ...interface{}) {
	t.errf(fmtString, args...)
	os.Exit(1)
}

func (t TerminalLogger) Print(args ...interface{}) {
	t.log(args...)
}

func (t TerminalLogger) Printf(fmtString string, args ...interface{}) {
	t.logf(fmtString, args...)
}

func (t TerminalLogger) Warn(args ...interface{}) {
	t.log(args...)
}

func (t TerminalLogger) Warnf(fmtString string, args ...interface{}) {
	t.logf(fmtString, args...)
}

func (t TerminalLogger) WithFields(fields Fields) Logger {
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
