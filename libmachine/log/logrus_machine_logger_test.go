package log

import (
	"testing"

	"bufio"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLevelIsInfo(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	assert.Equal(t, testLogger.Logger().Level, logrus.InfoLevel)
}

func TestSetDebugToTrue(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.SetDebug(true)
	assert.Equal(t, testLogger.Logger().Level, logrus.DebugLevel)
}

func TestSetDebugToFalse(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.SetDebug(true)
	testLogger.SetDebug(false)
	assert.Equal(t, testLogger.Logger().Level, logrus.InfoLevel)
}

func TestSetSilenceOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger().(*LogrusMachineLogger)
	testLogger.RedirectStdOutToStdErr()
	assert.Equal(t, testLogger.Logger().Level, logrus.ErrorLevel)
}

func TestDebugOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()
	testLogger.SetDebug(true)

	result := captureOutput(testLogger, func() { testLogger.Debug("debug") })

	assert.Equal(t, result, "debug")
}

func TestInfoOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Info("info") })

	assert.Equal(t, result, "info")
}

func TestWarnOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Warn("warn") })

	assert.Equal(t, result, "warn")
}

func TestErrorOutput(t *testing.T) {
	testLogger := NewLogrusMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Error("error") })

	assert.Equal(t, result, "error")
}

func TestEntriesAreCollected(t *testing.T) {
	testLogger := NewLogrusMachineLogger()
	testLogger.RedirectStdOutToStdErr()
	testLogger.Debug("debug")
	testLogger.Info("info")
	testLogger.Error("error")
	assert.Equal(t, 3, len(testLogger.History()))
	assert.Equal(t, "debug", testLogger.History()[0])
	assert.Equal(t, "info", testLogger.History()[1])
	assert.Equal(t, "error", testLogger.History()[2])
}

func captureOutput(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetOutput(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}
