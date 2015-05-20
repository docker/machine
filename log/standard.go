package log

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type standardLogger struct {
	stdout io.Writer
	stderr io.Writer
	fields Fields
}

func newStandardLogger(stdout io.Writer, stderr io.Writer, fields Fields) *standardLogger {
	if stderr == nil {
		stderr = stdout
	}
	return &standardLogger{stdout, stderr, fields}
}

func (s *standardLogger) Debug(args ...interface{}) {
	if isDebug() {
		s.out(args...)
	}
}

func (s *standardLogger) Debugf(fmtString string, args ...interface{}) {
	if isDebug() {
		s.outf(fmtString, args...)
	}
}

func (s *standardLogger) Debugln(args ...interface{}) {
	if isDebug() {
		s.outln(args...)
	}
}

func (s *standardLogger) Error(args ...interface{}) {
	s.err(args...)
}

func (s *standardLogger) Errorf(fmtString string, args ...interface{}) {
	s.errf(fmtString, args...)
}

func (s *standardLogger) Errorln(args ...interface{}) {
	s.errln(args...)
}

func (s *standardLogger) Info(args ...interface{}) {
	s.out(args...)
}

func (s *standardLogger) Infof(fmtString string, args ...interface{}) {
	s.outf(fmtString, args...)
}

func (s *standardLogger) Infoln(args ...interface{}) {
	s.outln(args...)
}

func (s *standardLogger) Fatal(args ...interface{}) {
	s.err(args...)
	os.Exit(1)
}

func (s *standardLogger) Fatalf(fmtString string, args ...interface{}) {
	s.errf(fmtString, args...)
	os.Exit(1)
}

func (s *standardLogger) Fatalln(args ...interface{}) {
	s.errln(args...)
	os.Exit(1)
}

func (s *standardLogger) Print(args ...interface{}) {
	s.out(args...)
}

func (s *standardLogger) Printf(fmtString string, args ...interface{}) {
	s.outf(fmtString, args...)
}

func (s *standardLogger) Println(args ...interface{}) {
	s.outln(args...)
}

func (s *standardLogger) Warn(args ...interface{}) {
	s.out(args...)
}

func (s *standardLogger) Warnf(fmtString string, args ...interface{}) {
	s.outf(fmtString, args...)
}

func (s *standardLogger) Warnln(args ...interface{}) {
	s.outln(args...)
}

func (s *standardLogger) WithField(fieldName string, field interface{}) Logger {
	return s.WithFields(Fields{fieldName: field})
}

func (s *standardLogger) WithFields(fields Fields) Logger {
	c := copyFields(s.fields)
	for k, v := range fields {
		c[k] = v
	}
	return newStandardLogger(s.stdout, s.stderr, c)
}

func (s *standardLogger) out(args ...interface{}) {
	s.print(s.stdout, fmt.Sprint(args...), "")
}

func (s *standardLogger) outf(fmtString string, args ...interface{}) {
	s.print(s.stdout, fmt.Sprintf(fmtString, args...), "")
}

func (s *standardLogger) outln(args ...interface{}) {
	s.print(s.stdout, fmt.Sprint(args...), "\n")
}

func (s *standardLogger) err(args ...interface{}) {
	s.print(s.stderr, fmt.Sprint(args...), "")
}

func (s *standardLogger) errf(fmtString string, args ...interface{}) {
	s.print(s.stderr, fmt.Sprintf(fmtString, args...), "")
}

func (s *standardLogger) errln(args ...interface{}) {
	s.print(s.stderr, fmt.Sprint(args...), "\n")
}

func (s *standardLogger) print(writer io.Writer, value string, suffix string) {
	if len(s.fields) == 0 {
		writer.Write([]byte(fmt.Sprintf("%s%s", value, suffix)))
		return
	}

	kvpairs := []string{}

	// Why the string slice song and dance?  Because Go's map iteration
	// order is random, we will get inconsistent results if we don't sort
	// the fields (or their resulting string K/V pairs, like we have here).
	// Otherwise, we couldn't test this reliably.
	for k, v := range s.fields {
		kvpairs = append(kvpairs, fmt.Sprintf("%s=%v", k, v))
	}

	sort.Strings(kvpairs)

	// TODO: Add more tabs?
	writer.Write([]byte(fmt.Sprintf("%s\t\t %s%s", value, strings.Join(kvpairs, " "), suffix)))
}
