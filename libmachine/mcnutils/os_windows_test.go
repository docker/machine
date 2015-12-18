package mcnutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOSWindows(t *testing.T) {
	output := `

Microsoft Windows [version 6.3.9600]

`

	assert.Equal(t, "Microsoft Windows [version 6.3.9600]", parseOutput(output))
}
