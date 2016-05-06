package dockerclient

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/go-connections/sockets"
)

func newHTTPClient(u *url.URL, httpTransport *http.Transport, timeout time.Duration) (*http.Client, error) {
	switch u.Scheme {
	default:
		if httpTransport.Dial == nil {
			directDialer := &net.Dialer{
				Timeout: timeout,
			}

			proxyDialer, err := sockets.DialerFromEnvironment(directDialer)
			if err != nil {
				return nil, err
			}
			httpTransport.Dial = proxyDialer.Dial
		}
	case "unix":
		socketPath := u.Path
		unixDial := func(proto, addr string) (net.Conn, error) {
			return net.DialTimeout("unix", socketPath, timeout)
		}
		httpTransport.Dial = unixDial
		// Override the main URL object so the HTTP lib won't complain
		u.Scheme = "http"
		u.Host = "unix.sock"
		u.Path = ""
	}
	return &http.Client{Transport: httpTransport}, nil
}
