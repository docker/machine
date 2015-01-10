package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libtrust"
	"github.com/docker/machine/drivers"
)

var (
	validHostNameChars   = `[a-zA-Z0-9_]`
	validHostNamePattern = regexp.MustCompile(`^` + validHostNameChars + `+$`)
)

type Host struct {
	Name       string `json:"-"`
	DriverName string
	Driver     drivers.Driver
	storePath  string
}

type hostConfig struct {
	DriverName string
}

func waitForDocker(addr string) error {
	for {
		log.Debugf("checking if Docker is available on %s", addr)
		conn, err := net.DialTimeout("tcp", addr, time.Second*5)
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}
		conn.Close()
		break
	}
	return nil
}

func NewHost(name, driverName, storePath string) (*Host, error) {
	driver, err := drivers.NewDriver(driverName, storePath)
	if err != nil {
		return nil, err
	}
	return &Host{
		Name:       name,
		DriverName: driverName,
		Driver:     driver,
		storePath:  storePath,
	}, nil
}

func LoadHost(name string, storePath string) (*Host, error) {
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}

	host := &Host{Name: name, storePath: storePath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}
	return host, nil
}

func ValidateHostName(name string) (string, error) {
	if !validHostNamePattern.MatchString(name) {
		return name, fmt.Errorf("Invalid host name %q, it must match %s", name, validHostNamePattern)
	}
	return name, nil
}

func loadTrustKey(trustKeyPath string) (libtrust.PrivateKey, error) {
	if err := os.MkdirAll(filepath.Dir(trustKeyPath), 0700); err != nil {
		return nil, err

	}

	trustKey, err := libtrust.LoadKeyFile(trustKeyPath)
	if err == libtrust.ErrKeyFileDoesNotExist {
		trustKey, err = libtrust.GenerateECP256PrivateKey()
		if err != nil {
			return nil, fmt.Errorf("error generating key: %s", err)
		}

		if err := libtrust.SaveKey(trustKeyPath, trustKey); err != nil {
			return nil, fmt.Errorf("error saving key file: %s", err)

		}

		dir, file := filepath.Split(trustKeyPath)
		if err := libtrust.SavePublicKey(filepath.Join(dir, "public-"+file), trustKey.PublicKey()); err != nil {
			return nil, fmt.Errorf("error saving public key file: %s", err)

		}
	} else if err != nil {
		return nil, fmt.Errorf("error loading key file: %s", err)

	}
	return trustKey, nil
}

func (h *Host) addHostToKnownHosts() error {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS10,
	}

	trustKeyPath := filepath.Join(drivers.GetDockerDir(), "key.json")
	knownHostsPath := filepath.Join(drivers.GetDockerDir(), "known-hosts.json")

	addr, err := h.GetDockerAddr()
	if err != nil {
		return err
	}

	if addr == "" {
		return nil
	}

	proto := "tcp"

	trustKey, err := loadTrustKey(trustKeyPath)
	if err != nil {
		return fmt.Errorf("unable to load trust key: %s", err)
	}

	knownHosts, err := libtrust.LoadKeySetFile(knownHostsPath)
	if err != nil {
		return fmt.Errorf("could not load trusted hosts file: %s", err)
	}

	allowedHosts, err := libtrust.FilterByHosts(knownHosts, addr, false)
	if err != nil {
		return fmt.Errorf("error filtering hosts: %s", err)
	}

	certPool, err := libtrust.GenerateCACertPool(trustKey, allowedHosts)
	if err != nil {
		return fmt.Errorf("Could not create CA pool: %s", err)
	}

	tlsConfig.ServerName = "docker"
	tlsConfig.RootCAs = certPool

	x509Cert, err := libtrust.GenerateSelfSignedClientCert(trustKey)
	if err != nil {
		return fmt.Errorf("certificate generation error: %s", err)
	}

	tlsConfig.Certificates = []tls.Certificate{{
		Certificate: [][]byte{x509Cert.Raw},
		PrivateKey:  trustKey.CryptoPrivateKey(),
		Leaf:        x509Cert,
	}}

	tlsConfig.InsecureSkipVerify = true

	testConn, err := tls.Dial(proto, addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls Handshake error: %s", err)
	}

	opts := x509.VerifyOptions{
		Roots:         tlsConfig.RootCAs,
		CurrentTime:   time.Now(),
		DNSName:       tlsConfig.ServerName,
		Intermediates: x509.NewCertPool(),
	}

	certs := testConn.ConnectionState().PeerCertificates
	for i, cert := range certs {
		if i == 0 {
			continue
		}
		opts.Intermediates.AddCert(cert)
	}

	if _, err := certs[0].Verify(opts); err != nil {
		if _, ok := err.(x509.UnknownAuthorityError); ok {
			pubKey, err := libtrust.FromCryptoPublicKey(certs[0].PublicKey)
			if err != nil {
				return fmt.Errorf("error extracting public key from cert: %s", err)
			}

			pubKey.AddExtendedField("hosts", []string{addr})

			log.Debugf("Adding machine to known hosts: %s", addr)

			if err := libtrust.AddKeySetFile(knownHostsPath, pubKey); err != nil {
				return fmt.Errorf("error adding machine to known hosts: %s", err)
			}
		}
	}

	testConn.Close()
	return nil
}

func (h *Host) Create() error {
	if err := h.Driver.Create(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	addr, err := h.GetDockerAddr()
	if err != nil {
		return err
	}

	log.Info("Waiting for machine to be provisioned and available. This may take a few minutes...")

	if addr != "" {
		// wait for provisioning before checking; otherwise
		// some providers will block the host due to the number of requests
		time.Sleep(30 * time.Second)
		if err := waitForDocker(addr); err != nil {
			return err
		}

		if err := h.addHostToKnownHosts(); err != nil {
			return err
		}

	}

	return nil
}

func (h *Host) Start() error {
	return h.Driver.Start()
}

func (h *Host) Stop() error {
	return h.Driver.Stop()
}

func (h *Host) Upgrade() error {
	return h.Driver.Upgrade()
}

func (h *Host) Remove(force bool) error {
	if err := h.Driver.Remove(); err != nil {
		if force {
			log.Errorf("Error removing host, force removing anyway: %s", err)
		} else {
			return err
		}
	}
	return h.removeStorePath()
}

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.storePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.storePath)
	}
	return os.RemoveAll(h.storePath)
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) GetIP() (string, error) {
	return h.Driver.GetIP()
}

func (h *Host) GetDockerAddr() (string, error) {
	driverUrl, err := h.GetURL()
	if err != nil {
		return "", fmt.Errorf("unable to get machine url: %s", err)
	}

	u, err := url.Parse(driverUrl)
	if err != nil {
		return "", fmt.Errorf("unable to parse machine url")
	}

	if u.Scheme == "unix" {
		return "", nil
	}

	return u.Host, nil
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.storePath, "config.json"))
	if err != nil {
		return err
	}

	// First pass: find the driver name and load the driver
	var config hostConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	driver, err := drivers.NewDriver(config.DriverName, h.storePath)
	if err != nil {
		return err
	}
	h.Driver = driver

	// Second pass: unmarshal driver config into correct driver
	if err := json.Unmarshal(data, &h); err != nil {
		return err
	}

	return nil
}

func (h *Host) SaveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(h.storePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}
