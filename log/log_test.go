package log

import "testing"

func TestTerminalLoggerWithFields(t *testing.T) {
	logger := TerminalLogger{}
	withFieldsLogger := logger.WithFields(Fields{
		"foo":  "bar",
		"spam": "eggs",
	})
	withFieldsTerminalLogger, ok := withFieldsLogger.(TerminalLogger)
	if !ok {
		t.Fatal("Type assertion to TerminalLogger failed")
	}
	expectedOutFields := "\t\t foo=bar spam=eggs"
	if withFieldsTerminalLogger.fieldOut != expectedOutFields {
		t.Fatalf("Expected %q, got %q", expectedOutFields, withFieldsTerminalLogger.fieldOut)
	}
}
