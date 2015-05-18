package log

import (
	"bytes"
	"testing"
)

func TestStandardLoggerWithFields(t *testing.T) {
	buffer := bytes.NewBuffer(nil)

	logger := newStandardLogger(buffer, buffer, make(Fields))
	withFieldsLogger := logger.WithFields(Fields{
		"spam": "eggs",
		"foo":  "bar",
	})
	withFieldsLogger.Println("woot")
	logger.Print("hello")

	expected := "woot\t\t foo=bar spam=eggs\nhello"
	actual := buffer.String()
	if expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}
