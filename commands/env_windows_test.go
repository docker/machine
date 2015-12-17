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
