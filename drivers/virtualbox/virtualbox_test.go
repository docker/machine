package virtualbox

import (
	"net"
	"reflect"
	"testing"
)

func TestGetRandomIPinSubnet(t *testing.T) {
	// test IP 1.2.3.4
	testIP := net.IPv4(byte(1), byte(2), byte(3), byte(4))
	newIP, err := getRandomIPinSubnet(testIP)
	if err != nil {
		t.Fatal(err)
	}

	if testIP.Equal(newIP) {
		t.Fatalf("expected different IP (source %s); received %s", testIP.String(), newIP.String())
	}

	if newIP[0] != testIP[0] {
		t.Fatalf("expected first octet of %d; received %d", testIP[0], newIP[0])
	}

	if newIP[1] != testIP[1] {
		t.Fatalf("expected second octet of %d; received %d", testIP[1], newIP[1])
	}

	if newIP[2] != testIP[2] {
		t.Fatalf("expected third octet of %d; received %d", testIP[2], newIP[2])
	}
}

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

	n, err := getHostOnlyNetwork(vboxNets, ip, ipnet.Mask)
	if err != nil {
		t.Fatalf("Was not expecting an error calling getHostOnlyNetwork: %s", err)
	}

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

	n, err := getHostOnlyNetwork(vboxNets, ip, ipnet.Mask)
	if err == nil {
		t.Fatalf("Expected to get an error but instead got no error")
	}
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
	n, err := getHostOnlyNetwork(vboxNets, ip, net.IPMask(net.ParseIP("255.255.255.0").To4()))
	if err != nil {
		t.Fatalf("Was not expecting an error calling getHostOnlyNetwork: %s", err)
	}

	if !reflect.DeepEqual(n, expectedHostOnlyNetwork) {
		t.Fatalf("Expected result of calling getHostOnlyNetwork to be the same as expected but it was not:\nexpected: %+v\nactual: %+v\n", expectedHostOnlyNetwork, n)
	}
}
