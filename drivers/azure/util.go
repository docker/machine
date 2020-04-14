package azure

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/rancher/machine/drivers/azure/azureutil"
	"github.com/rancher/machine/drivers/azure/logutil"
	"github.com/rancher/machine/drivers/driverutil"
	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/libmachine/ssh"
	"github.com/rancher/machine/libmachine/state"
)

var (
	supportedEnvironments = []string{
		azure.PublicCloud.Name,
		azure.USGovernmentCloud.Name,
		azure.ChinaCloud.Name,
		azure.GermanCloud.Name,
	}
)

// requiredOptionError forms an error from the error indicating the option has
// to be provided with a value for this driver.
type requiredOptionError string

func (r requiredOptionError) Error() string {
	return fmt.Sprintf("%s driver requires the %q option.", driverName, string(r))
}

// newAzureClient creates an AzureClient helper from the Driver context and
// initiates authentication if required.
func (d *Driver) newAzureClient(ctx context.Context) (*azureutil.AzureClient, error) {
	env, err := azure.EnvironmentFromName(d.Environment)
	if err != nil {
		supportedValues := strings.Join(supportedEnvironments, ", ")
		return nil, fmt.Errorf("Invalid Azure environment: %q, supported values: %s", d.Environment, supportedValues)
	}

	var (
		authorizer *autorest.BearerAuthorizer
	)
	if d.ClientID != "" && d.ClientSecret != "" { // use client credentials auth
		log.Debug("Using Azure client credentials.")
		authorizer, err = azureutil.AuthenticateClientCredentials(ctx, env, d.SubscriptionID, d.ClientID, d.ClientSecret)
		if err != nil {
			return nil, fmt.Errorf("Failed to authenticate using client credentials: %+v", err)
		}
	} else { // use browser-based device auth
		log.Debug("Using Azure device flow authentication.")
		authorizer, err = azureutil.AuthenticateDeviceFlow(ctx, env, d.SubscriptionID)
		if err != nil {
			return nil, fmt.Errorf("Error creating Azure client: %v", err)
		}
	}
	return azureutil.New(env, d.SubscriptionID, authorizer), nil
}

// generateSSHKey creates a ssh key pair locally and saves the public key file
// contents in OpenSSH format to the DeploymentContext.
func (d *Driver) generateSSHKey(deploymentCtx *azureutil.DeploymentContext) error {
	privPath := d.GetSSHKeyPath()
	pubPath := privPath + ".pub"

	log.Debug("Creating SSH key...", logutil.Fields{
		"pub":  pubPath,
		"priv": privPath,
	})

	if err := ssh.GenerateSSHKey(privPath); err != nil {
		return err
	}
	log.Debug("SSH key pair generated.")

	publicKey, err := ioutil.ReadFile(pubPath)
	deploymentCtx.SSHPublicKey = string(publicKey)
	return err
}

// getSecurityRules creates network security group rules based on driver
// configuration such as SSH port, docker port and swarm port.
func (d *Driver) getSecurityRules(extraPorts []string) (*[]network.SecurityRule, error) {
	mkRule := func(priority int, name, description, srcPort, dstPort string, proto network.SecurityRuleProtocol) network.SecurityRule {
		return network.SecurityRule{
			Name: to.StringPtr(name),
			SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
				Description:              to.StringPtr(description),
				SourceAddressPrefix:      to.StringPtr("*"),
				DestinationAddressPrefix: to.StringPtr("*"),
				SourcePortRange:          to.StringPtr(srcPort),
				DestinationPortRange:     to.StringPtr(dstPort),
				Access:                   network.SecurityRuleAccessAllow,
				Direction:                network.SecurityRuleDirectionInbound,
				Protocol:                 proto,
				Priority:                 to.Int32Ptr(int32(priority)),
			},
		}
	}

	log.Debugf("Docker port is configured as %d", d.DockerPort)

	// Base ports to be opened for any machine
	rl := []network.SecurityRule{
		mkRule(100, "SSHAllowAny", "Allow ssh from public Internet", "*", fmt.Sprintf("%d", d.BaseDriver.SSHPort), network.SecurityRuleProtocolTCP),
		mkRule(300, "DockerAllowAny", "Allow docker engine access (TLS-protected)", "*", fmt.Sprintf("%d", d.DockerPort), network.SecurityRuleProtocolTCP),
	}

	// extra port numbers requested by user
	basePri := 1000
	for i, p := range extraPorts {
		port, protocol := driverutil.SplitPortProto(p)
		proto, err := parseSecurityRuleProtocol(protocol)
		if err != nil {
			return nil, fmt.Errorf("cannot parse security rule protocol: %v", err)
		}
		log.Debugf("User-requested port to be opened on NSG: %v/%s", port, proto)
		name := fmt.Sprintf("Port%s-%sAllowAny", port, proto)
		name = strings.Replace(name, "*", "Asterisk", -1)
		r := mkRule(basePri+i, name, "User requested port to be accessible from Internet via docker-machine", "*", port, proto)
		rl = append(rl, r)
	}
	log.Debugf("Total NSG rules: %d", len(rl))

	return &rl, nil
}

func (d *Driver) naming() azureutil.ResourceNaming {
	return azureutil.ResourceNaming(d.BaseDriver.MachineName)
}

// ipAddress returns machine’s private or public IP address according to the
// configuration. If no IP address is found it returns empty string.
func (d *Driver) ipAddress(ctx context.Context) (ip string, err error) {
	c, err := d.newAzureClient(ctx)
	if err != nil {
		return "", err
	}

	var ipType string
	if d.UsePrivateIP || d.NoPublicIP {
		ipType = "Private"
		ip, err = c.GetPrivateIPAddress(ctx, d.ResourceGroup, d.naming().NIC())
	} else {
		ipType = "Public"
		ip, err = c.GetPublicIPAddress(ctx, d.ResourceGroup,
			d.naming().IP(),
			d.DNSLabel != "")
	}

	log.Debugf("Retrieving %s IP address...", ipType)
	if err != nil {
		return "", fmt.Errorf("Error querying %s IP: %v", ipType, err)
	}
	if ip == "" {
		log.Debugf("%s IP address is not yet allocated.", ipType)
	}
	return ip, nil
}

func machineStateForVMPowerState(ps azureutil.VMPowerState) state.State {
	m := map[azureutil.VMPowerState]state.State{
		azureutil.Running:      state.Running,
		azureutil.Starting:     state.Starting,
		azureutil.Stopping:     state.Stopping,
		azureutil.Stopped:      state.Stopped,
		azureutil.Deallocating: state.Stopping,
		azureutil.Deallocated:  state.Stopped,
		azureutil.Unknown:      state.None,
	}

	if v, ok := m[ps]; ok {
		return v
	}
	log.Warnf("Azure PowerState %q does not map to a docker-machine state.", ps)
	return state.None
}

// parseVirtualNetwork parses Virtual Network input format "[resourcegroup:]name"
// into Resource Group (uses provided one if omitted) and Virtual Network Name
func parseVirtualNetwork(name string, defaultRG string) (string, string) {
	l := strings.SplitN(name, ":", 2)
	if len(l) == 2 {
		return l[0], l[1]
	}
	return defaultRG, name
}

// parseSecurityRuleProtocol parses a protocol string into a network.SecurityRuleProtocol
// and returns error if the protocol is not supported
func parseSecurityRuleProtocol(proto string) (network.SecurityRuleProtocol, error) {
	switch strings.ToLower(proto) {
	case "tcp":
		return network.SecurityRuleProtocolTCP, nil
	case "udp":
		return network.SecurityRuleProtocolUDP, nil
	case "*":
		return network.SecurityRuleProtocolAsterisk, nil
	default:
		return "", fmt.Errorf("invalid protocol %s", proto)
	}
}
