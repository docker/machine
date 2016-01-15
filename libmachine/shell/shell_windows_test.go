package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetect(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	shell, err := Detect()

	assert.Equal(t, "cmd", shell)
	assert.NoError(t, err)
}

func TestStartedBy(t *testing.T) {
	shell, err := startedBy()

	assert.Equal(t, "go.exe", shell)
	assert.NoError(t, err)
}
