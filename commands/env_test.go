package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHints(t *testing.T) {
	var tests = []struct {
		userShell string
		noProxy   bool
		swarm     bool
		hints     string
	}{
		{"", false, false, "# Run this command to configure your shell: \n# eval \"$(machine env default)\"\n"},
		{"", true, false, "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy default)\"\n"},
		{"", false, true, "# Run this command to configure your shell: \n# eval \"$(machine env --swarm default)\"\n"},
		{"", true, true, "# Run this command to configure your shell: \n# eval \"$(machine env --no-proxy --swarm default)\"\n"},

		{"fish", false, false, "# Run this command to configure your shell: \n# eval (machine env --shell=fish default)\n"},
		{"fish", true, false, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --no-proxy default)\n"},
		{"fish", false, true, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --swarm default)\n"},
		{"fish", true, true, "# Run this command to configure your shell: \n# eval (machine env --shell=fish --no-proxy --swarm default)\n"},

		{"powershell", false, false, "# Run this command to configure your shell: \n# machine env --shell=powershell default | Invoke-Expression\n"},
		{"powershell", true, false, "# Run this command to configure your shell: \n# machine env --shell=powershell --no-proxy default | Invoke-Expression\n"},
		{"powershell", false, true, "# Run this command to configure your shell: \n# machine env --shell=powershell --swarm default | Invoke-Expression\n"},
		{"powershell", true, true, "# Run this command to configure your shell: \n# machine env --shell=powershell --no-proxy --swarm default | Invoke-Expression\n"},

		{"cmd", false, false, "# Run this command to configure your shell: \n# copy and paste the above values into your command prompt\n"},
		{"cmd", true, false, "# Run this command to configure your shell: \n# copy and paste the above values into your command prompt\n"},
		{"cmd", false, true, "# Run this command to configure your shell: \n# copy and paste the above values into your command prompt\n"},
		{"cmd", true, true, "# Run this command to configure your shell: \n# copy and paste the above values into your command prompt\n"},
	}

	for _, expected := range tests {
		hints := generateUsageHint("machine", "default", expected.userShell, expected.noProxy, expected.swarm)

		assert.Equal(t, expected.hints, hints)
	}
}
