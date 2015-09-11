package firewallaction

type FirewallAction int

const (
	Enable FirewallAction = iota
	Disable
	Allow
	Deny
	Reload
	Unload
	Save
)
