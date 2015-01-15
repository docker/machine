package store

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
	"github.com/docker/machine/state"
)

var (
	validHostNameChars   = `[a-zA-Z0-9_]`
	validHostNamePattern = regexp.MustCompile(`^` + validHostNameChars + `+$`)
)

type Host struct {
	Name       string `json:"-"`
	DriverName string
	Driver     drivers.Driver
	StorePath  string
}

type hostConfig struct {
	DriverName string
}

func waitForDocker(addr string) error {
	for {
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

func newHost(name, driverName, storePath string) (*Host, error) {
	driver, err := drivers.NewDriver(driverName, name, storePath)
	if err != nil {
		return nil, err
	}
	return &Host{
		Name:       name,
		DriverName: driverName,
		Driver:     driver,
		StorePath:  storePath,
	}, nil
}

func loadHost(name string, storePath string) (*Host, error) {
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}

	host := &Host{Name: name, StorePath: storePath}
	if err := host.loadConfig(); err != nil {
		return nil, err
	}
	return host, nil
}

func validateHostName(name string) (string, error) {
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

	driverUrl, err := h.GetURL()
	if err != nil {
		return fmt.Errorf("unable to get machine url: %s", err)
	}

	u, err := url.Parse(driverUrl)
	if err != nil {
		return fmt.Errorf("unable to parse machine url")
	}

	if u.Scheme == "unix" {
		return nil
	}

	addr := u.Host
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

	log.Debugf("waiting for Docker to become available on %s", addr)
	if err := waitForDocker(addr); err != nil {
		return fmt.Errorf("unable to connect to Docker daemon: %s", err)
	}
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

func (h *Host) create(name string) error {
	if err := h.Driver.Create(); err != nil {
		return err
	}

	if err := h.saveConfig(); err != nil {
		return err
	}

	if err := h.addHostToKnownHosts(); err != nil {
		return err
	}

	return nil
}

func (h *Host) start() error {
	return h.Driver.Start()
}

func (h *Host) stop() error {
	return h.Driver.Stop()
}

func (h *Host) upgrade() error {
	return h.Driver.Upgrade()
}

func (h *Host) remove(force bool) error {
	if err := h.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}
	return h.removeStorePath()
}

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.StorePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.StorePath)
	}
	return os.RemoveAll(h.StorePath)
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

func (h *Host) GetIP() (string, error) {
	return h.Driver.GetIP()
}

func (h *Host) GetState() (state.State, error) {
	return h.Driver.GetState()
}

func (h *Host) GetDriverName() string {
	return h.Driver.DriverName()
}

func (h *Host) loadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.StorePath, "config.json"))
	if err != nil {
		return err
	}

	// First pass: find the driver name and load the driver
	var config hostConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	driver, err := drivers.NewDriver(config.DriverName, h.Name, h.StorePath)
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

func (h *Host) saveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(h.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}
