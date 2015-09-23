package host

import (
	"os"
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

func TestMigrateHostV0ToV1(t *testing.T) {
	os.Setenv("MACHINE_STORAGE_PATH", "/tmp/migration")
	originalHost := &HostV0{
		HostOptions:    nil,
		SwarmDiscovery: "token://foobar",
		SwarmHost:      "1.2.3.4:2376",
		SwarmMaster:    true,
		CaCertPath:     "/tmp/migration/certs/ca.pem",
		PrivateKeyPath: "/tmp/migration/certs/ca-key.pem",
		ClientCertPath: "/tmp/migration/certs/cert.pem",
		ClientKeyPath:  "/tmp/migration/certs/key.pem",
		ServerCertPath: "/tmp/migration/certs/server.pem",
		ServerKeyPath:  "/tmp/migration/certs/server-key.pem",
	}
	hostOptions := &HostOptionsV1{
		SwarmOptions: &swarm.SwarmOptions{
			Master:    true,
			Discovery: "token://foobar",
			Host:      "1.2.3.4:2376",
		},
		AuthOptions: &AuthOptionsV1{
			CaCertPath:     "/tmp/migration/certs/ca.pem",
			PrivateKeyPath: "/tmp/migration/certs/ca-key.pem",
			ClientCertPath: "/tmp/migration/certs/cert.pem",
			ClientKeyPath:  "/tmp/migration/certs/key.pem",
			ServerCertPath: "/tmp/migration/certs/server.pem",
			ServerKeyPath:  "/tmp/migration/certs/server-key.pem",
		},
		EngineOptions: &engine.EngineOptions{},
	}

	expectedHost := &HostV1{
		HostOptions: hostOptions,
	}

	host := MigrateHostV0ToHostV1(originalHost)

	if !reflect.DeepEqual(host, expectedHost) {
		t.Logf("\n%+v\n%+v", host, expectedHost)
		t.Logf("\n%+v\n%+v", host.HostOptions, expectedHost.HostOptions)
		t.Fatal("Expected these structs to be equal, they were different")
	}
}

func TestMigrateHostMetadataV0ToV1(t *testing.T) {
	metadata := &HostMetadataV0{
		HostOptions: HostOptions{
			EngineOptions: nil,
			AuthOptions:   nil,
		},
		StorePath:      "/tmp/store",
		CaCertPath:     "/tmp/store/certs/ca.pem",
		ServerCertPath: "/tmp/store/certs/server.pem",
	}
	expectedAuthOptions := &auth.AuthOptions{
		StorePath:      "/tmp/store",
		CaCertPath:     "/tmp/store/certs/ca.pem",
		ServerCertPath: "/tmp/store/certs/server.pem",
	}

	expectedMetadata := &HostMetadata{
		HostOptions: HostOptions{
			EngineOptions: &engine.EngineOptions{},
			AuthOptions:   expectedAuthOptions,
		},
	}

	m := MigrateHostMetadataV0ToHostMetadataV1(metadata)

	if !reflect.DeepEqual(m, expectedMetadata) {
		t.Logf("\n%+v\n%+v", m, expectedMetadata)
		t.Fatal("Expected these structs to be equal, they were different")
	}
}
