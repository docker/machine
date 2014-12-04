package virtualbox

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	reHostonlyInterfaceCreated = regexp.MustCompile(`Interface '(.+)' was successfully created`)
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

// Config changes the configuration of the host-only network.
func (n *hostOnlyNetwork) Save() error {
	if n.IPv4.IP != nil && n.IPv4.Mask != nil {
		if err := vbm("hostonlyif", "ipconfig", n.Name, "--ip", n.IPv4.IP.String(), "--netmask", net.IP(n.IPv4.Mask).String()); err != nil {
			return err
		}
	}

	if n.IPv6.IP != nil && n.IPv6.Mask != nil {
		prefixLen, _ := n.IPv6.Mask.Size()
		if err := vbm("hostonlyif", "ipconfig", n.Name, "--ipv6", n.IPv6.IP.String(), "--netmasklengthv6", fmt.Sprintf("%d", prefixLen)); err != nil {
			return err
		}
	}

	if n.DHCP {
		vbm("hostonlyif", "ipconfig", n.Name, "--dhcp") // not implemented as of VirtualBox 4.3
	}

	return nil
}

// createHostonlyNet creates a new host-only network.
func createHostonlyNet() (*hostOnlyNetwork, error) {
	out, err := vbmOut("hostonlyif", "create")
	if err != nil {
		return nil, err
	}
	res := reHostonlyInterfaceCreated.FindStringSubmatch(string(out))
	if res == nil {
		return nil, errors.New("failed to create hostonly interface")
	}
	return &hostOnlyNetwork{Name: res[1]}, nil
}

// HostonlyNets gets all host-only networks in a  map keyed by HostonlyNet.NetworkName.
func listHostOnlyNetworks() (map[string]*hostOnlyNetwork, error) {
	out, err := vbmOut("list", "hostonlyifs")
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(strings.NewReader(out))
	m := map[string]*hostOnlyNetwork{}
	n := &hostOnlyNetwork{}
	for s.Scan() {
		line := s.Text()
		if line == "" {
			m[n.NetworkName] = n
			n = &hostOnlyNetwork{}
			continue
		}
		res := reColonLine.FindStringSubmatch(line)
		if res == nil {
			continue
		}
		switch key, val := res[1], res[2]; key {
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
			l, err := strconv.ParseUint(val, 10, 7)
			if err != nil {
				return nil, err
			}
			n.IPv6.Mask = net.CIDRMask(int(l), net.IPv6len*8)
		case "HardwareAddress":
			mac, err := net.ParseMAC(val)
			if err != nil {
				return nil, err
			}
			n.HwAddr = mac
		case "MediumType":
			n.Medium = val
		case "Status":
			n.Status = val
		case "VBoxNetworkName":
			n.NetworkName = val
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

func getHostOnlyNetwork(hostIP net.IP, netmask net.IPMask) (*hostOnlyNetwork, error) {
	nets, err := listHostOnlyNetworks()
	if err != nil {
		return nil, err
	}
	for _, n := range nets {
		if hostIP.Equal(n.IPv4.IP) &&
			netmask.String() == n.IPv4.Mask.String() {
			return n, nil
		}
	}
	return nil, nil
}

func getOrCreateHostOnlyNetwork(hostIP net.IP, netmask net.IPMask, dhcpIP net.IP, dhcpUpperIP net.IP, dhcpLowerIP net.IP) (*hostOnlyNetwork, error) {
	hostOnlyNet, err := getHostOnlyNetwork(hostIP, netmask)
	if err != nil || hostOnlyNet != nil {
		return hostOnlyNet, err
	}
	// No existing host-only interface found. Create a new one.
	hostOnlyNet, err = createHostonlyNet()
	if err != nil {
		return nil, err
	}
	hostOnlyNet.IPv4.IP = hostIP
	hostOnlyNet.IPv4.Mask = netmask
	if err := hostOnlyNet.Save(); err != nil {
		return nil, err
	}

	dhcp := dhcpServer{}
	dhcp.IPv4.IP = dhcpIP
	dhcp.IPv4.Mask = netmask
	dhcp.LowerIP = dhcpUpperIP
	dhcp.UpperIP = dhcpLowerIP
	dhcp.Enabled = true
	if err := addHostonlyDHCP(hostOnlyNet.Name, dhcp); err != nil {
		return nil, err
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

func addDHCPServer(kind, name string, d dhcpServer) error {
	command := "modify"

	// On some platforms (OSX), creating a hostonlyinterface adds a default dhcpserver
	// While on others (Windows?) it does not.
	dhcps, err := getDHCPServers()
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
	return vbm(args...)
}

// AddInternalDHCP adds a DHCP server to an internal network.
func addInternalDHCP(netname string, d dhcpServer) error {
	return addDHCPServer("--netname", netname, d)
}

// AddHostonlyDHCP adds a DHCP server to a host-only network.
func addHostonlyDHCP(ifname string, d dhcpServer) error {
	return addDHCPServer("--netname", "HostInterfaceNetworking-"+ifname, d)
}

// DHCPs gets all DHCP server settings in a map keyed by DHCP.NetworkName.
func getDHCPServers() (map[string]*dhcpServer, error) {
	out, err := vbmOut("list", "dhcpservers")
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(strings.NewReader(out))
	m := map[string]*dhcpServer{}
	dhcp := &dhcpServer{}
	for s.Scan() {
		line := s.Text()
		if line == "" {
			m[dhcp.NetworkName] = dhcp
			dhcp = &dhcpServer{}
			continue
		}
		res := reColonLine.FindStringSubmatch(line)
		if res == nil {
			continue
		}
		switch key, val := res[1], res[2]; key {
		case "NetworkName":
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
	}
	if err := s.Err(); err != nil {
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
