package cert

import (
	"os"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
)

func BootstrapCertificates(authOptions *auth.AuthOptions) error {
	certDir := authOptions.CertDir
	caCertPath := authOptions.CaCertPath
	caPrivateKeyPath := authOptions.CaPrivateKeyPath
	clientCertPath := authOptions.ClientCertPath
	clientKeyPath := authOptions.ClientKeyPath

	// TODO: I'm not super happy about this use of "org", the user should
	// have to specify it explicitly instead of implicitly basing it on
	// $USER.
	org := mcnutils.GetUsername()

	bits := 2048

	if _, err := os.Stat(certDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(certDir, 0700); err != nil {
				log.Fatalf("Error creating machine certificate dir: %s", err)
			}
		} else {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		log.Infof("Creating CA: %s", caCertPath)

		// check if the key path exists; if so, error
		if _, err := os.Stat(caPrivateKeyPath); err == nil {
			log.Fatalf("The CA key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := GenerateCACertificate(caCertPath, caPrivateKeyPath, org, bits); err != nil {
			log.Infof("Error generating CA certificate: %s", err)
		}
	}

	if _, err := os.Stat(clientCertPath); os.IsNotExist(err) {
		log.Infof("Creating client certificate: %s", clientCertPath)

		if _, err := os.Stat(certDir); err != nil {
			if os.IsNotExist(err) {
				if err := os.Mkdir(certDir, 0700); err != nil {
					log.Fatalf("Error creating machine client cert dir: %s", err)
				}
			} else {
				log.Fatal(err)
			}
		}

		// check if the key path exists; if so, error
		if _, err := os.Stat(clientKeyPath); err == nil {
			log.Fatalf("The client key already exists.  Please remove it or specify a different key/cert.")
		}

		if err := GenerateCert([]string{""}, clientCertPath, clientKeyPath, caCertPath, caPrivateKeyPath, org, bits); err != nil {
			log.Fatalf("Error generating client certificate: %s", err)
		}
	}

	return nil
}
