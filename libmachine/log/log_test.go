package log

import "testing"

func TestStandardLoggerWithFields(t *testing.T) {
	logger := StandardLogger{}
	withFieldsLogger := logger.WithFields(Fields{
		"foo":  "bar",
		"spam": "eggs",
	})
	withFieldsStandardLogger, ok := withFieldsLogger.(StandardLogger)
	if !ok {
		t.Fatal("Type assertion to StandardLogger failed")
	}
	expectedOutFields := "\t\t foo=bar spam=eggs"
	if withFieldsStandardLogger.fieldOut != expectedOutFields {
		t.Fatalf("Expected %q, got %q", expectedOutFields, withFieldsStandardLogger.fieldOut)
	}
}
