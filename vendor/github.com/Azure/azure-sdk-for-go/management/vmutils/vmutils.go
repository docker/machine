// Package vmutils provides convenience methods for creating Virtual
// Machine Role configurations.
package vmutils

import (
	"fmt"

	vm "github.com/Azure/azure-sdk-for-go/management/virtualmachine"
)

const (
	errParamNotSpecified = "Parameter %s is not specified."
)

// NewVMConfiguration creates configuration for a new virtual machine Role.
func NewVMConfiguration(name string, roleSize string) vm.Role {
	return vm.Role{
		RoleName:            name,
		RoleType:            "PersistentVMRole",
		RoleSize:            roleSize,
		ProvisionGuestAgent: true,
	}
}

// ConfigureForLinux adds configuration for when deploying a generalized Linux
// image. If "password" is left empty, SSH password security will be disabled by
// default. Certificates with SSH public keys should already be uploaded to the
// cloud service where the VM will be deployed and referenced here only by their
// thumbprint.
func ConfigureForLinux(role *vm.Role, hostname, user, password string, sshPubkeyCertificateThumbprint ...string) error {
	if role == nil {
		return fmt.Errorf(errParamNotSpecified, "role")
	}

	role.ConfigurationSets = updateOrAddConfig(role.ConfigurationSets, vm.ConfigurationSetTypeLinuxProvisioning,
		func(config *vm.ConfigurationSet) {
			config.HostName = hostname
			config.UserName = user
			config.UserPassword = password
			if password != "" {
				config.DisableSSHPasswordAuthentication = "false"
			}
			if len(sshPubkeyCertificateThumbprint) != 0 {
				config.SSH = &vm.SSH{}
				for _, k := range sshPubkeyCertificateThumbprint {
					config.SSH.PublicKeys = append(config.SSH.PublicKeys,
						vm.PublicKey{
							Fingerprint: k,
							Path:        "/home/" + user + "/.ssh/authorized_keys",
						},
					)
				}
			}
		},
	)

	return nil
}

// ConfigureForWindows adds configuration for when deploying a generalized
// Windows image. timeZone can be left empty. For a complete list of supported
// time zone entries, you can either refer to the values listed in the registry
// entry "HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows NT\CurrentVersion\Time
// Zones" or you can use the tzutil command-line tool to list the valid time.
func ConfigureForWindows(role *vm.Role, hostname, user, password string, enableAutomaticUpdates bool, timeZone string) error {
	if role == nil {
		return fmt.Errorf(errParamNotSpecified, "role")
	}

	role.ConfigurationSets = updateOrAddConfig(role.ConfigurationSets, vm.ConfigurationSetTypeWindowsProvisioning,
		func(config *vm.ConfigurationSet) {
			config.ComputerName = hostname
			config.AdminUsername = user
			config.AdminPassword = password
			config.EnableAutomaticUpdates = enableAutomaticUpdates
			config.TimeZone = timeZone
		},
	)

	return nil
}

// ConfigureWindowsToJoinDomain adds configuration to join a new Windows vm to a
// domain. "username" must be in UPN form (user@domain.com), "machineOU" can be
// left empty
func ConfigureWindowsToJoinDomain(role *vm.Role, username, password, domainToJoin, machineOU string) error {
	if role == nil {
		return fmt.Errorf(errParamNotSpecified, "role")
	}

	winconfig := findConfig(role.ConfigurationSets, vm.ConfigurationSetTypeWindowsProvisioning)
	if winconfig != nil {
		winconfig.DomainJoin = &vm.DomainJoin{
			Credentials:     vm.Credentials{Username: username, Password: password},
			JoinDomain:      domainToJoin,
			MachineObjectOU: machineOU,
		}
	}

	return nil
}

func ConfigureWinRMListener(role *vm.Role, protocol vm.WinRMProtocol, certificateThumbprint string) error {

	if role == nil {
		return fmt.Errorf(errParamNotSpecified, "role")
	}

	winconfig := findConfig(role.ConfigurationSets, vm.ConfigurationSetTypeWindowsProvisioning)

	if winconfig != nil {

		listener := vm.WinRMListener{
			Protocol:              protocol,
			CertificateThumbprint: certificateThumbprint,
		}

		if winconfig.WinRMListeners == nil {
			winconfig.WinRMListeners = &[]vm.WinRMListener{}
		}

		currentListeners := *winconfig.WinRMListeners

		// replace existing listener if it's already configured
		for i, existingListener := range currentListeners {
			if existingListener.Protocol == protocol {
				currentListeners[i] = listener
				return nil
			}
		}

		// otherwise append to list of listeners
		newListeners := append(currentListeners, listener)
		winconfig.WinRMListeners = &newListeners

		return nil
	}

	return fmt.Errorf("WindowsProvisioningConfigurationSet not found in 'role'")
}

func ConfigureWinRMOverHTTP(role *vm.Role) error {
	return ConfigureWinRMListener(role, vm.WinRMProtocolHTTP, "")
}

func ConfigureWinRMOverHTTPS(role *vm.Role, certificateThumbprint string) error {
	return ConfigureWinRMListener(role, vm.WinRMProtocolHTTPS, certificateThumbprint)
}
