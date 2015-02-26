package utils

import (
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var passphrases = [][]byte{
	{}, // This will produce an unencrypted CA key file
	[]byte("foobar"),
}

func TestGenerateCACertificate(t *testing.T) {
	for _, pass := range passphrases {
		tmpDir, err := ioutil.TempDir("", "machine-test-")
		if err != nil {
			t.Fatal(err)
		}

		os.Setenv("MACHINE_DIR", tmpDir)
		caCertPath := filepath.Join(tmpDir, "ca.pem")
		caKeyPath := filepath.Join(tmpDir, "key.pem")
		testOrg := "test-org"
		bits := 2048
		if err := GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits, pass); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(caCertPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(caKeyPath); err != nil {
			t.Fatal(err)
		}
		os.Setenv("MACHINE_DIR", "")

		// cleanup
		_ = os.RemoveAll(tmpDir)
	}
}

func TestGenerateCert(t *testing.T) {
	for _, pass := range passphrases {

		// First we must override getpasswd in certs.go to return the expected
		// passphrase bytes.
		getpasswd = func() []byte {
			return pass
		}

		tmpDir, err := ioutil.TempDir("", "machine-test-")
		if err != nil {
			t.Fatal(err)
		}

		os.Setenv("MACHINE_DIR", tmpDir)
		caCertPath := filepath.Join(tmpDir, "ca.pem")
		caKeyPath := filepath.Join(tmpDir, "key.pem")
		certPath := filepath.Join(tmpDir, "cert.pem")
		keyPath := filepath.Join(tmpDir, "cert-key.pem")
		testOrg := "test-org"
		bits := 2048
		if err := GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits, pass); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(caCertPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(caKeyPath); err != nil {
			t.Fatal(err)
		}
		os.Setenv("MACHINE_DIR", "")

		if err := GenerateCert([]string{}, certPath, keyPath, caCertPath, caKeyPath, testOrg, bits); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(certPath); err != nil {
			t.Fatalf("certificate not created at %s", certPath)
		}

		if _, err := os.Stat(keyPath); err != nil {
			t.Fatalf("key not created at %s", keyPath)
		}

		// cleanup
		_ = os.RemoveAll(tmpDir)
	}
}

func TestGenerateCertWithInvalidPassphrase(t *testing.T) {
	passwd := []byte("foobar")
	invalids := [][]byte{
		{},
		[]byte("testing"),
	}

	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	// Always clean up
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	for _, invalid := range invalids {
		getpasswd = func() []byte {
			return invalid
		}

		os.Setenv("MACHINE_DIR", tmpDir)
		caCertPath := filepath.Join(tmpDir, "ca.pem")
		caKeyPath := filepath.Join(tmpDir, "key.pem")
		certPath := filepath.Join(tmpDir, "cert.pem")
		keyPath := filepath.Join(tmpDir, "cert-key.pem")
		testOrg := "test-org"
		bits := 2048
		if err := GenerateCACertificate(caCertPath, caKeyPath, testOrg, bits, passwd); err != nil {
			t.Fatal(err)
		}

		if _, err := os.Stat(caCertPath); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(caKeyPath); err != nil {
			t.Fatal(err)
		}
		os.Setenv("MACHINE_DIR", "")

		if err := GenerateCert([]string{}, certPath, keyPath, caCertPath, caKeyPath, testOrg, bits); err != x509.IncorrectPasswordError {
			t.Fatal(err)
		}

	}

}
