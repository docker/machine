package commands

import (
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
)

func TestHints(t *testing.T) {
	var tests = []struct {
		userShell     string
		commandLine   string
		expectedHints string
	}{
		{"", "machine env default", "# Run this command to configure your shell: \n# eval \"$(machine env default)\"\n"},
		{"", "machine env --no-proxy default", "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy default)\"\n"},
		{"", "machine env --swarm default", "# Run this command to configure your shell: \n# eval \"$(machine env --swarm default)\"\n"},
		{"", "machine env --no-proxy --swarm default", "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy --swarm default)\"\n"},

		{"fish", "./machine env --shell=fish default", "# Run this command to configure your shell: \n# eval (./machine env --shell=fish default)\n"},
		{"fish", "./machine env --shell=fish --no-proxy default", "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --no-proxy default)\n"},
		{"fish", "./machine env --shell=fish --swarm default", "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --swarm default)\n"},
		{"fish", "./machine env --shell=fish --no-proxy --swarm default", "# Run this command to configure your shell: \n# eval (./machine env --shell=fish --no-proxy --swarm default)\n"},

		{"powershell", "./machine env --shell=powershell default", "# Run this command to configure your shell: \n# ./machine env --shell=powershell default | Invoke-Expression\n"},
		{"powershell", "./machine env --shell=powershell --no-proxy default", "# Run this command to configure your shell: \n# ./machine env --shell=powershell --no-proxy default | Invoke-Expression\n"},
		{"powershell", "./machine env --shell=powershell --swarm default", "# Run this command to configure your shell: \n# ./machine env --shell=powershell --swarm default | Invoke-Expression\n"},
		{"powershell", "./machine env --shell=powershell --no-proxy --swarm default", "# Run this command to configure your shell: \n# ./machine env --shell=powershell --no-proxy --swarm default | Invoke-Expression\n"},

		{"cmd", "./machine env --shell=cmd default", "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd default') DO %i\n"},
		{"cmd", "./machine env --shell=cmd --no-proxy default", "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --no-proxy default') DO %i\n"},
		{"cmd", "./machine env --shell=cmd --swarm default", "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --swarm default') DO %i\n"},
		{"cmd", "./machine env --shell=cmd --no-proxy --swarm default", "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('./machine env --shell=cmd --no-proxy --swarm default') DO %i\n"},
	}

	for _, test := range tests {
		hints := generateUsageHint(test.userShell, strings.Split(test.commandLine, " "))

		assert.Equal(t, test.expectedHints, hints)
	}
}
