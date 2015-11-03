package commands

import (
	"errors"
	"testing"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/stretchr/testify/assert"
)

type FakeValidateCertificate struct {
	IsValid bool
	Err     error
}

type FakeCertGenerator struct {
	fakeValidateCertificate *FakeValidateCertificate
}

func (fcg FakeCertGenerator) GenerateCACertificate(certFile, keyFile, org string, bits int) error {
	return nil
}

func (fcg FakeCertGenerator) GenerateCert(hosts []string, certFile, keyFile, caFile, caKeyFile, org string, bits int) error {
	return nil
}

func (fcg FakeCertGenerator) ValidateCertificate(addr string, authOptions *auth.AuthOptions) (bool, error) {
	return fcg.fakeValidateCertificate.IsValid, fcg.fakeValidateCertificate.Err
}

func TestCheckCert(t *testing.T) {
	errCertsExpired := errors.New("Certs have expired")

	cases := []struct {
		hostUrl     string
		authOptions *auth.AuthOptions
		valid       bool
		checkErr    error
		expectedErr error
	}{
		{"192.168.99.100:2376", &auth.AuthOptions{}, true, nil, nil},
		{"192.168.99.100:2376", &auth.AuthOptions{}, false, nil, ErrCertInvalid{wrappedErr: nil, hostUrl: "192.168.99.100:2376"}},
		{"192.168.99.100:2376", &auth.AuthOptions{}, false, errCertsExpired, ErrCertInvalid{wrappedErr: errCertsExpired, hostUrl: "192.168.99.100:2376"}},
	}

	for _, c := range cases {
		fcg := FakeCertGenerator{fakeValidateCertificate: &FakeValidateCertificate{c.valid, c.checkErr}}
		cert.SetCertGenerator(fcg)
		err := checkCert(c.hostUrl, c.authOptions)
		assert.Equal(t, c.expectedErr, err)
	}
}
