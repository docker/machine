package commands

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBash(t *testing.T) {
	originalShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "/bin/bash")
	defer os.Setenv("SHELL", originalShell)
	shell, err := detectShell()
	assert.Nil(t, err)
	assert.Equal(t, "bash", shell)
}

func TestDetectFish(t *testing.T) {
	originalShell := os.Getenv("SHELL")
	os.Setenv("SHELL", "/bin/bash")
	defer os.Setenv("SHELL", originalShell)
	originalFishdir := os.Getenv("__fish_bin_dir")
	os.Setenv("__fish_bin_dir", "/usr/local/Cellar/fish/2.2.0/bin")
	defer os.Setenv("__fish_bin_dir", originalFishdir)
	shell, err := detectShell()
	assert.Nil(t, err)
	assert.Equal(t, "fish", shell)
}
