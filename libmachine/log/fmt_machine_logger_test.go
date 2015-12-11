package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFmtDebug(t *testing.T) {
	testLogger := NewFmtMachineLogger()
	testLogger.SetDebug(true)

	result := captureOutput(testLogger, func() { testLogger.Debug("debug") })

	assert.Equal(t, result, "debug")
}

func TestFmtInfo(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Info("info") })

	assert.Equal(t, result, "info")
}

func TestFmtWarn(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Warn("warn") })

	assert.Equal(t, result, "warn")
}

func TestFmtError(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Error("error") })

	assert.Equal(t, result, "error")
}
