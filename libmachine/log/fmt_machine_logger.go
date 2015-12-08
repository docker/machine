package log

import (
	"fmt"
	"io"
	"os"
)

type FmtMachineLogger struct {
	out   io.Writer
	err   io.Writer
	debug bool
}

// NewFmtMachineLogger creates a MachineLogger implementation used by the drivers
func NewFmtMachineLogger() MachineLogger {
	return &FmtMachineLogger{
		out:   os.Stdout,
		err:   os.Stderr,
		debug: false,
	}
}

func (ml *FmtMachineLogger) RedirectStdOutToStdErr() {
	ml.out = ml.err
}

func (ml *FmtMachineLogger) SetDebug(debug bool) {
	ml.debug = debug
}

func (ml *FmtMachineLogger) SetOutput(out io.Writer) {
	ml.out = out
	ml.err = out
}

func (ml *FmtMachineLogger) Debug(args ...interface{}) {
	if ml.debug {
		fmt.Fprintln(ml.err, args...)
	}
}

func (ml *FmtMachineLogger) Debugf(fmtString string, args ...interface{}) {
	if ml.debug {
		fmt.Fprintf(ml.err, fmtString+"\n", args...)
	}
}

func (ml *FmtMachineLogger) Error(args ...interface{}) {
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Errorf(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Info(args ...interface{}) {
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Infof(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Fatal(args ...interface{}) {
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Fatalf(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Warn(args ...interface{}) {
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Warnf(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) History() []string {
	return []string{}
}
