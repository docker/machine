package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	originalShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "")
	defer os.Setenv("SHELL", originalShell)
	shell, err := detectShell()
	assert.Nil(t, err)
	assert.Equal(t, "cmd", shell)
}

func TestStartedBy(t *testing.T) {
	shell, err := startedBy()
	assert.Nil(t, err)
	assert.NotNil(t, shell)
	assert.Equal(t, "go.exe", shell)
}
