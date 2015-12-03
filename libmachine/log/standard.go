package log

import (
	"fmt"
	"io"
	"os"
	"sync"
)

type StandardLogger struct {
	OutWriter io.Writer
	ErrWriter io.Writer
	mu        *sync.Mutex
}

func (t StandardLogger) log(args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprint(t.OutWriter, args...)
	fmt.Fprint(t.OutWriter, "\n")
}

func (t StandardLogger) logf(fmtString string, args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprintf(t.OutWriter, fmtString, args...)
	fmt.Fprint(t.OutWriter, "\n")
}

func (t StandardLogger) err(args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprint(t.ErrWriter, args...)
	fmt.Fprint(t.ErrWriter, "\n")
}

func (t StandardLogger) errf(fmtString string, args ...interface{}) {
	defer t.mu.Unlock()
	t.mu.Lock()
	fmt.Fprintf(t.ErrWriter, fmtString, args...)
	fmt.Fprint(t.ErrWriter, "\n")
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

func (t StandardLogger) Warn(args ...interface{}) {
	fmt.Print("WARNING >>> ")
	t.log(args...)
}

func (t StandardLogger) Warnf(fmtString string, args ...interface{}) {
	fmt.Print("WARNING >>> ")
	t.logf(fmtString, args...)
}
