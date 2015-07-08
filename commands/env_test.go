package commands

import (
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
)

type MockCmdEnvFlags struct {
	flags string
}

func (f *MockCmdEnvFlags) Bool(name string) bool {
	return strings.Contains(f.flags, name)
}

func TestHints(t *testing.T) {
	var tests = []struct {
		userShell     string
		flags         *MockCmdEnvFlags
		expectedHints string
	}{
		{"", &MockCmdEnvFlags{}, "# Run this command to configure your shell: \n# eval \"$(machine env default)\"\n"},
		{"", &MockCmdEnvFlags{"--no-proxy"}, "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy default)\"\n"},
		{"", &MockCmdEnvFlags{"--swarm"}, "# Run this command to configure your shell: \n# eval \"$(machine env --swarm default)\"\n"},
		{"", &MockCmdEnvFlags{"--no-proxy --swarm"}, "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy --swarm default)\"\n"},

		{"fish", &MockCmdEnvFlags{}, "# Run this command to configure your shell: \n# eval (machine env --shell=fish default)\n"},
		{"fish", &MockCmdEnvFlags{"--no-proxy"}, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --no-proxy default)\n"},
		{"fish", &MockCmdEnvFlags{"--swarm"}, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --swarm default)\n"},
		{"fish", &MockCmdEnvFlags{"--no-proxy --swarm"}, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --no-proxy --swarm default)\n"},

		{"powershell", &MockCmdEnvFlags{}, "# Run this command to configure your shell: \n# machine env --shell=powershell default | Invoke-Expression\n"},
		{"powershell", &MockCmdEnvFlags{"--no-proxy"}, "# Run this command to configure your shell: \n# machine env --shell=powershell --no-proxy default | Invoke-Expression\n"},
		{"powershell", &MockCmdEnvFlags{"--swarm"}, "# Run this command to configure your shell: \n# machine env --shell=powershell --swarm default | Invoke-Expression\n"},
		{"powershell", &MockCmdEnvFlags{"--no-proxy --swarm"}, "# Run this command to configure your shell: \n# machine env --shell=powershell --no-proxy --swarm default | Invoke-Expression\n"},

		{"cmd", &MockCmdEnvFlags{}, "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('machine env --shell=cmd default') DO %i\n"},
		{"cmd", &MockCmdEnvFlags{"--no-proxy"}, "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('machine env --shell=cmd --no-proxy default') DO %i\n"},
		{"cmd", &MockCmdEnvFlags{"--swarm"}, "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('machine env --shell=cmd --swarm default') DO %i\n"},
		{"cmd", &MockCmdEnvFlags{"--no-proxy --swarm"}, "REM Run this command to configure your shell: \nREM \tFOR /f \"tokens=*\" %i IN ('machine env --shell=cmd --no-proxy --swarm default') DO %i\n"},
	}

	for _, test := range tests {
		hints := generateUsageHint("machine", "default", test.userShell, test.flags)

		assert.Equal(t, test.expectedHints, hints)
	}
}
