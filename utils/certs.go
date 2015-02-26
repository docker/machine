package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"time"

	"github.com/howeyc/gopass"
)

func newCertificate(org string) (*x509.Certificate, error) {
	now := time.Now()
	// need to set notBefore slightly in the past to account for time
	// skew in the VMs otherwise the certs sometimes are not yet valid
	notBefore := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute()-5, 0, 0, time.Local)
	notAfter := notBefore.Add(time.Hour * 24 * 1080)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	return &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{org},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}, nil

}

// GenerateCACertificate generates a new certificate authority from the specified org
// and bit size and stores the resulting certificate and key file
// in the arguments.
func GenerateCACertificate(certFile, keyFile, org string, bits int, optPassphrase []byte) error {
	var pemBlock *pem.Block

	template, err := newCertificate(org)
	if err != nil {
		return err
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err

	}

	pemBlock = &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}
	// Encrypt the private key using the supplied passphrase
	if len(optPassphrase) > 0 {
		pemBlock, err = x509.EncryptPEMBlock(rand.Reader, "RSA PRIVATE KEY", pemBlock.Bytes, optPassphrase, x509.PEMCipherAES256)
		if err != nil {
			return err
		}
	}

	pem.Encode(keyOut, pemBlock)
	keyOut.Close()

	return nil
}

// GenerateCert generates a new certificate signed using the provided
// certificate authority files and stores the result in the certificate
// file and key provided.  The provided host names are set to the
// appropriate certificate fields.
func GenerateCert(hosts []string, certFile, keyFile, caFile, caKeyFile, org string, bits int) error {
	template, err := newCertificate(org)
	if err != nil {
		return err
	}
	// client
	if len(hosts) == 1 && hosts[0] == "" {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
		template.KeyUsage = x509.KeyUsageDigitalSignature
	} else { // server
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}
		for _, h := range hosts {
			if ip := net.ParseIP(h); ip != nil {
				template.IPAddresses = append(template.IPAddresses, ip)

			} else {
				template.DNSNames = append(template.DNSNames, h)
			}
		}
	}

	tlsCert, err := LoadX509KeyPair(caFile, caKeyFile)
	if err != nil {
		return err
	}

	priv, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err

	}

	x509Cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
	if err != nil {
		return err
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, template, x509Cert, &priv.PublicKey, tlsCert.PrivateKey)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err

	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err

	}

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()

	return nil
}

// We use an external library, gopass, to prompt the user to enter their
// password on the command line for encrypted CA key files.
//
// However, in testing this method is unavailable. By assigning this method
// to a variable we can change the variable in testing to a []byte array
// matching the password. This allows us to automate testing of encrypted certs
var getpasswd = gopass.GetPasswd

// Attempts to load an encrypted X509 certificate
// If the key file is not encrypted it falls back to the default tls package
// implementation which works for unencrypted key files.
func LoadX509KeyPair(caFile, caKeyFile string) (cert tls.Certificate, err error) {
	var caByt, encKeyByt []byte

	if encKeyByt, err = ioutil.ReadFile(caKeyFile); err != nil {
		return
	}

	pemBlock, _ := pem.Decode(encKeyByt)
	if pemBlock == nil {
		return cert, fmt.Errorf("Unable to parse certificate key file")
	}

	if !x509.IsEncryptedPEMBlock(pemBlock) {
		// This is unencrypted, therefore we can use the noraml tls package implementation
		return tls.LoadX509KeyPair(caFile, caKeyFile)
	}

	// Decrypt the key with a given password
	fmt.Printf("CA Key password: ")

	// In normal environments getpasswd calls gopass.Getpasswd; in testing we override
	// this variable to return passwords without asking for input.
	if pemBlock.Bytes, err = x509.DecryptPEMBlock(pemBlock, getpasswd()); err != nil {
		return
	}

	if caByt, err = ioutil.ReadFile(caFile); err != nil {
		return
	}

	// We can now return a standard X509 key pair using the decrypted PEM key in memory
	return tls.X509KeyPair(caByt, pem.EncodeToMemory(pemBlock))
}
