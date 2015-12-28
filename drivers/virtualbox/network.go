package virtualbox

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
)

const (
	buggyNetmask = "0f000000"
)

var (
	reHostonlyInterfaceCreated        = regexp.MustCompile(`Interface '(.+)' was successfully created`)
	errNewHostOnlyInterfaceNotVisible = errors.New("The host-only interface we just created is not visible. This is a well known bug of VirtualBox. You might want to uninstall it and reinstall at least version 5.0.12 that is is supposed to fix this issue")
)

// Host-only network.
type hostOnlyNetwork struct {
	Name        string
	GUID        string
	DHCP        bool
	IPv4        net.IPNet
	IPv6        net.IPNet
	HwAddr      net.HardwareAddr
	Medium      string
	Status      string
	NetworkName string // referenced in DHCP.NetworkName
}

// Save changes the configuration of the host-only network.
func (n *hostOnlyNetwork) Save(vbox VBoxManager) error {
	if n.IPv4.IP != nil && n.IPv4.Mask != nil {
		if err := vbox.vbm("hostonlyif", "ipconfig", n.Name, "--ip", n.IPv4.IP.String(), "--netmask", net.IP(n.IPv4.Mask).String()); err != nil {
			return err
		}
	}

	if n.IPv6.IP != nil && n.IPv6.Mask != nil {
		prefixLen, _ := n.IPv6.Mask.Size()
		if err := vbox.vbm("hostonlyif", "ipconfig", n.Name, "--ipv6", n.IPv6.IP.String(), "--netmasklengthv6", fmt.Sprintf("%d", prefixLen)); err != nil {
			return err
		}
	}

	if n.DHCP {
		vbox.vbm("hostonlyif", "ipconfig", n.Name, "--dhcp") // not implemented as of VirtualBox 4.3
	}

	return nil
}

// createHostonlyNet creates a new host-only network.
func createHostonlyNet(vbox VBoxManager) (*hostOnlyNetwork, error) {
	out, err := vbox.vbmOut("hostonlyif", "create")
	if err != nil {
		return nil, err
	}

	res := reHostonlyInterfaceCreated.FindStringSubmatch(string(out))
	if res == nil {
		return nil, errors.New("failed to create hostonly interface")
	}

	return &hostOnlyNetwork{Name: res[1]}, nil
}

// listHostOnlyNetworks gets all host-only networks in a  map keyed by HostonlyNet.NetworkName.
func listHostOnlyNetworks(vbox VBoxManager) (map[string]*hostOnlyNetwork, error) {
	out, err := vbox.vbmOut("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}

	byName := map[string]*hostOnlyNetwork{}
	byIP := map[string]*hostOnlyNetwork{}
	n := &hostOnlyNetwork{}

	err = parseKeyValues(out, reColonLine, func(key, val string) error {
		switch key {
		case "Name":
			n.Name = val
		case "GUID":
			n.GUID = val
		case "DHCP":
			n.DHCP = (val != "Disabled")
		case "IPAddress":
			n.IPv4.IP = net.ParseIP(val)
		case "NetworkMask":
			n.IPv4.Mask = parseIPv4Mask(val)
		case "IPV6Address":
			n.IPv6.IP = net.ParseIP(val)
		case "IPV6NetworkMaskPrefixLength":
			l, err := strconv.ParseUint(val, 10, 8)
			if err != nil {
				return err
			}
			n.IPv6.Mask = net.CIDRMask(int(l), net.IPv6len*8)
		case "HardwareAddress":
			mac, err := net.ParseMAC(val)
			if err != nil {
				return err
			}
			n.HwAddr = mac
		case "MediumType":
			n.Medium = val
		case "Status":
			n.Status = val
		case "VBoxNetworkName":
			n.NetworkName = val

			if _, present := byName[n.NetworkName]; present {
				return fmt.Errorf("VirtualBox is configured with multiple host-only interfaces with the same name %q. Please remove one.", n.NetworkName)
			}
			byName[n.NetworkName] = n

			if len(n.IPv4.IP) != 0 {
				if _, present := byIP[n.IPv4.IP.String()]; present {
					return fmt.Errorf("VirtualBox is configured with multiple host-only interfaces with the same IP %q. Please remove one.", n.IPv4.IP)
				}
				byIP[n.IPv4.IP.String()] = n
			}

			n = &hostOnlyNetwork{}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return byName, nil
}

func getHostOnlyNetwork(nets map[string]*hostOnlyNetwork, hostIP net.IP, netmask net.IPMask) *hostOnlyNetwork {
	for _, n := range nets {
		// Second part of this conditional handles a race where
		// VirtualBox returns us the incorrect netmask value for the
		// newly created interface.
		if hostIP.Equal(n.IPv4.IP) &&
			(netmask.String() == n.IPv4.Mask.String() || n.IPv4.Mask.String() == buggyNetmask) {
			return n
		}
	}

	return nil
}

func getOrCreateHostOnlyNetwork(hostIP net.IP, netmask net.IPMask, dhcpIP net.IP, dhcpLowerIP net.IP, dhcpUpperIP net.IP, vbox VBoxManager) (*hostOnlyNetwork, error) {
	nets, err := listHostOnlyNetworks(vbox)
	if err != nil {
		return nil, err
	}

	hostOnlyNet := getHostOnlyNetwork(nets, hostIP, netmask)
	if hostOnlyNet != nil {
		return hostOnlyNet, nil
	}

	// No existing host-only interface found. Create a new one.
	hostOnlyNet, err = createHostonlyNet(vbox)
	if err != nil {
		return nil, err
	}

	hostOnlyNet.IPv4.IP = hostIP
	hostOnlyNet.IPv4.Mask = netmask
	if err := hostOnlyNet.Save(vbox); err != nil {
		return nil, err
	}

	dhcp := dhcpServer{}
	dhcp.IPv4.IP = dhcpIP
	dhcp.IPv4.Mask = netmask
	dhcp.LowerIP = dhcpLowerIP
	dhcp.UpperIP = dhcpUpperIP
	dhcp.Enabled = true
	if err := addHostonlyDHCP(hostOnlyNet.Name, dhcp, vbox); err != nil {
		return nil, err
	}

	// Check that the interface really exists.
	// Sometimes, Vbox says it created the interface but then it cannot be found...
	nets, err = listHostOnlyNetworks(vbox)
	if err != nil {
		return nil, err
	}
	hostOnlyNet = getHostOnlyNetwork(nets, hostIP, netmask)
	if hostOnlyNet == nil {
		return nil, errNewHostOnlyInterfaceNotVisible
	}

	return hostOnlyNet, nil
}

// DHCP server info.
type dhcpServer struct {
	NetworkName string
	IPv4        net.IPNet
	LowerIP     net.IP
	UpperIP     net.IP
	Enabled     bool
}

func addDHCPServer(kind, name string, d dhcpServer, vbox VBoxManager) error {
	command := "modify"

	// On some platforms (OSX), creating a hostonlyinterface adds a default dhcpserver
	// While on others (Windows?) it does not.
	dhcps, err := getDHCPServers(vbox)
	if err != nil {
		return err
	}

	if _, ok := dhcps[name]; !ok {
		command = "add"
	}

	args := []string{"dhcpserver", command,
		kind, name,
		"--ip", d.IPv4.IP.String(),
		"--netmask", net.IP(d.IPv4.Mask).String(),
		"--lowerip", d.LowerIP.String(),
		"--upperip", d.UpperIP.String(),
	}
	if d.Enabled {
		args = append(args, "--enable")
	} else {
		args = append(args, "--disable")
	}

	return vbox.vbm(args...)
}

// addHostonlyDHCP adds a DHCP server to a host-only network.
func addHostonlyDHCP(ifname string, d dhcpServer, vbox VBoxManager) error {
	return addDHCPServer("--netname", "HostInterfaceNetworking-"+ifname, d, vbox)
}

// getDHCPServers gets all DHCP server settings in a map keyed by DHCP.NetworkName.
func getDHCPServers(vbox VBoxManager) (map[string]*dhcpServer, error) {
	out, err := vbox.vbmOut("list", "dhcpservers")
	if err != nil {
		return nil, err
	}

	m := map[string]*dhcpServer{}
	dhcp := &dhcpServer{}

	err = parseKeyValues(out, reColonLine, func(key, val string) error {
		switch key {
		case "NetworkName":
			dhcp = &dhcpServer{}
			m[val] = dhcp
			dhcp.NetworkName = val
		case "IP":
			dhcp.IPv4.IP = net.ParseIP(val)
		case "upperIPAddress":
			dhcp.UpperIP = net.ParseIP(val)
		case "lowerIPAddress":
			dhcp.LowerIP = net.ParseIP(val)
		case "NetworkMask":
			dhcp.IPv4.Mask = parseIPv4Mask(val)
		case "Enabled":
			dhcp.Enabled = (val == "Yes")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// parseIPv4Mask parses IPv4 netmask written in IP form (e.g. 255.255.255.0).
// This function should really belong to the net package.
func parseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}
