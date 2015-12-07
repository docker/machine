package log

import (
	"testing"

	"bufio"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLevelIsInfo(t *testing.T) {
	testLogger := NewMachineLogger()
	assert.Equal(t, testLogger.Logger().(*logrus.Logger).Level, logrus.InfoLevel)
}

func TestSetDebugToTrue(t *testing.T) {
	testLogger := NewMachineLogger()
	testLogger.SetDebug(true)
	assert.Equal(t, testLogger.Logger().(*logrus.Logger).Level, logrus.DebugLevel)
}

func TestSetDebugToFalse(t *testing.T) {
	testLogger := NewMachineLogger()
	testLogger.SetDebug(true)
	testLogger.SetDebug(false)
	assert.Equal(t, testLogger.Logger().(*logrus.Logger).Level, logrus.InfoLevel)
}

func TestSetSilenceOutput(t *testing.T) {
	testLogger := NewMachineLogger()
	testLogger.RedirectStdOutToStdErr()
	assert.Equal(t, testLogger.Logger().(*logrus.Logger).Level, logrus.ErrorLevel)
}

func TestDebug(t *testing.T) {
	testLogger := NewMachineLogger()
	testLogger.SetDebug(true)

	result := captureOutput(testLogger, func() { testLogger.Debug("debug") })

	assert.Equal(t, result, "debug")
}

func TestInfo(t *testing.T) {
	testLogger := NewMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Info("info") })

	assert.Equal(t, result, "info")
}

func TestWarn(t *testing.T) {
	testLogger := NewMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Warn("warn") })

	assert.Equal(t, result, "warn")
}

func TestError(t *testing.T) {
	testLogger := NewMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Error("error") })

	assert.Equal(t, result, "error")
}

func captureOutput(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetOutput(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}
