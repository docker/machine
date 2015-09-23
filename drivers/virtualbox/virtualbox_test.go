package virtualbox

import (
	"net"
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
