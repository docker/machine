package log

import (
	"io"
	"regexp"
)

const redactedText = "<REDACTED>"

var (
	Logger MachineLogger

	// (?s) enables '.' to match '\n' -- see https://golang.org/pkg/regexp/syntax/
	certRegex = regexp.MustCompile("(?s)-----BEGIN CERTIFICATE-----.*-----END CERTIFICATE-----")
	keyRegex  = regexp.MustCompile("(?s)-----BEGIN RSA PRIVATE KEY-----.*-----END RSA PRIVATE KEY-----")
)

func init() {
	Logger = NewFmtMachineLogger()
}

func stripSecrets(original []string) []string {
	stripped := []string{}
	for _, line := range original {
		line = certRegex.ReplaceAllString(line, redactedText)
		line = keyRegex.ReplaceAllString(line, redactedText)
		stripped = append(stripped, line)
	}
	return stripped
}

// RedirectStdOutToStdErr prevents any log from corrupting the output
func RedirectStdOutToStdErr() {
	Logger.RedirectStdOutToStdErr()
}

func Debug(args ...interface{}) {
	Logger.Debug(args...)
}

func Debugf(fmtString string, args ...interface{}) {
	Logger.Debugf(fmtString, args...)
}

func Error(args ...interface{}) {
	Logger.Error(args...)
}

func Errorf(fmtString string, args ...interface{}) {
	Logger.Errorf(fmtString, args...)
}

func Info(args ...interface{}) {
	Logger.Info(args...)
}

func Infof(fmtString string, args ...interface{}) {
	Logger.Infof(fmtString, args...)
}

func Fatal(args ...interface{}) {
	Logger.Fatal(args...)
}

func Fatalf(fmtString string, args ...interface{}) {
	Logger.Fatalf(fmtString, args...)
}

func Warn(args ...interface{}) {
	Logger.Warn(args...)
}

func Warnf(fmtString string, args ...interface{}) {
	Logger.Warnf(fmtString, args...)
}

func SetDebug(debug bool) {
	Logger.SetDebug(debug)
}

func SetOutput(out io.Writer) {
	Logger.SetOutput(out)
}

func History() []string {
	return stripSecrets(Logger.History())
}
