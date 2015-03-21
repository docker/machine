package libmachine

import (
	"os"
	"reflect"
	"testing"
)

// Tests a function which "prefills" certificate information for a host
// due to a schema migration from "flat" to a "nested" structure.
func TestGetCertInfoFromHost(t *testing.T) {
	os.Setenv("MACHINE_STORAGE_PATH", "/tmp/migration")
	host := &Host{
		CaCertPath:     "",
		PrivateKeyPath: "",
		ClientCertPath: "",
		ClientKeyPath:  "",
		ServerCertPath: "",
		ServerKeyPath:  "",
	}
	expectedCertInfo := CertPathInfo{
		CaCertPath:     "/tmp/migration/certs/ca.pem",
		CaKeyPath:      "/tmp/migration/certs/ca-key.pem",
		ClientCertPath: "/tmp/migration/certs/cert.pem",
		ClientKeyPath:  "/tmp/migration/certs/key.pem",
		ServerCertPath: "/tmp/migration/certs/server.pem",
		ServerKeyPath:  "/tmp/migration/certs/server-key.pem",
	}
	certInfo := getCertInfoFromHost(host)
	if !reflect.DeepEqual(expectedCertInfo, certInfo) {
		t.Log("\n\n\n", expectedCertInfo, "\n\n\n", certInfo)
		t.Fatal("Expected these structs to be equal, they were different")
	}
}
