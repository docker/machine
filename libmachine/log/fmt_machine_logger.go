package log

import (
	"fmt"
	"io"
	"os"
)

type FmtMachineLogger struct {
	out     io.Writer
	err     io.Writer
	debug   bool
	history *HistoryRecorder
}

// NewFmtMachineLogger creates a MachineLogger implementation used by the drivers
func NewFmtMachineLogger() MachineLogger {
	return &FmtMachineLogger{
		out:     os.Stdout,
		err:     os.Stderr,
		debug:   false,
		history: NewHistoryRecorder(),
	}
}

func (ml *FmtMachineLogger) SetDebug(debug bool) {
	ml.debug = debug
}

func (ml *FmtMachineLogger) SetOut(out io.Writer) {
	ml.out = out
}

func (ml *FmtMachineLogger) SetErr(err io.Writer) {
	ml.err = err
}

func (ml *FmtMachineLogger) Debug(args ...interface{}) {
	ml.history.Record(args...)
	if ml.debug {
		fmt.Fprintln(ml.err, args...)
	}
}

func (ml *FmtMachineLogger) Debugf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	if ml.debug {
		fmt.Fprintf(ml.err, fmtString+"\n", args...)
	}
}

func (ml *FmtMachineLogger) Error(args ...interface{}) {
	ml.history.Record(args...)
	fmt.Fprintln(ml.err, args...)
}

func (ml *FmtMachineLogger) Errorf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	fmt.Fprintf(ml.err, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Info(args ...interface{}) {
	ml.history.Record(args...)
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Infof(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Fatal(args ...interface{}) {
	ml.history.Record(args...)
	fmt.Fprintln(ml.err, args...)
	os.Exit(1)
}

func (ml *FmtMachineLogger) Fatalf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	fmt.Fprintf(ml.err, fmtString+"\n", args...)
	os.Exit(1)
}

func (ml *FmtMachineLogger) Warn(args ...interface{}) {
	ml.history.Record(args...)
	fmt.Fprintln(ml.out, args...)
}

func (ml *FmtMachineLogger) Warnf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	fmt.Fprintf(ml.out, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) History() []string {
	return ml.history.records
}
