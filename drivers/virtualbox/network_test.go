package virtualbox

import (
	"net"
	"reflect"
	"testing"
)

// Tests that when we have a host only network which matches our expectations,
// it gets returned correctly.
func TestGetHostOnlyNetworkHappy(t *testing.T) {
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}
	expectedHostOnlyNetwork := &hostOnlyNetwork{
		IPv4: *ipnet,
	}
	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": expectedHostOnlyNetwork,
	}

	n := getHostOnlyNetwork(vboxNets, ip, ipnet.Mask)
	if !reflect.DeepEqual(n, expectedHostOnlyNetwork) {
		t.Fatalf("Expected result of calling getHostOnlyNetwork to be the same as expected but it was not:\nexpected: %+v\nactual: %+v\n", expectedHostOnlyNetwork, n)
	}
}

// Tests that we are able to properly detect when a host only network which
// matches our expectations can not be found.
func TestGetHostOnlyNetworkNotFound(t *testing.T) {
	// Note that this has a different ip is different from "ip" below.
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}

	ip = net.ParseIP("192.168.59.0").To4()

	// Suppose a vbox net is created, but it doesn't align with our
	// expectation.
	vboxNet := &hostOnlyNetwork{
		IPv4: *ipnet,
	}
	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": vboxNet,
	}

	n := getHostOnlyNetwork(vboxNets, ip, ipnet.Mask)
	if n != nil {
		t.Fatalf("Expected vbox net to be nil but it has a value: %+v\n", n)
	}
}

// Tests a special case where Virtualbox creates the host only network
// successfully but mis-reports the netmask.
func TestGetHostOnlyNetworkWindows10Bug(t *testing.T) {
	cidr := "192.168.99.0/24"
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("Error parsing cidr: %s", err)
	}

	// This is a faulty netmask: a VirtualBox bug causes it to be
	// misreported.
	ipnet.Mask = net.IPMask(net.ParseIP("15.0.0.0").To4())

	expectedHostOnlyNetwork := &hostOnlyNetwork{
		IPv4: *ipnet,
	}

	vboxNets := map[string]*hostOnlyNetwork{
		"HostInterfaceNetworking-vboxnet0": expectedHostOnlyNetwork,
	}

	// The Mask that we are passing in will be the "legitimate" mask, so it
	// must differ from the magic buggy mask.
	n := getHostOnlyNetwork(vboxNets, ip, net.IPMask(net.ParseIP("255.255.255.0").To4()))
	if !reflect.DeepEqual(n, expectedHostOnlyNetwork) {
		t.Fatalf("Expected result of calling getHostOnlyNetwork to be the same as expected but it was not:\nexpected: %+v\nactual: %+v\n", expectedHostOnlyNetwork, n)
	}
}
