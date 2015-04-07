package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	timeout = time.Second * 5
)

func defaultTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func getClient() *http.Client {
	transport := http.Transport{
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
		Dial:              defaultTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}

	return &client
}

type B2dUtils struct {
	isoFilename      string
	commonIsoPath    string
	imgCachePath     string
	githubApiBaseUrl string
	githubBaseUrl    string
}

func NewB2dUtils(githubApiBaseUrl, githubBaseUrl string) *B2dUtils {
	defaultBaseApiUrl := "https://api.github.com"
	defaultBaseUrl := "https://github.com"
	imgCachePath := GetMachineCacheDir()
	isoFilename := "boot2docker.iso"

	if githubApiBaseUrl == "" {
		githubApiBaseUrl = defaultBaseApiUrl
	}

	if githubBaseUrl == "" {
		githubBaseUrl = defaultBaseUrl
	}

	return &B2dUtils{
		isoFilename:      isoFilename,
		imgCachePath:     GetMachineCacheDir(),
		commonIsoPath:    filepath.Join(imgCachePath, isoFilename),
		githubApiBaseUrl: githubApiBaseUrl,
		githubBaseUrl:    githubBaseUrl,
	}
}

// Get the latest boot2docker release tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func (b *B2dUtils) GetLatestBoot2DockerReleaseURL() (string, error) {
	client := getClient()
	apiUrl := fmt.Sprintf("%s/repos/boot2docker/boot2docker/releases", b.githubApiBaseUrl)
	rsp, err := client.Get(apiUrl)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	var t []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
		return "", err
	}
	if len(t) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	tag := t[0].TagName
	isoUrl := fmt.Sprintf("%s/boot2docker/boot2docker/releases/download/%s/boot2docker.iso", b.githubBaseUrl, tag)
	return isoUrl, nil
}

// Download boot2docker ISO image for the given tag and save it at dest.
func (b *B2dUtils) DownloadISO(dir, file, isoUrl string) error {
	u, err := url.Parse(isoUrl)
	var src io.ReadCloser
	if u.Scheme == "file" || u.Scheme == "" {
		s, err := os.Open(u.Path)
		if err != nil {
			return err
		}
		src = s
	} else {
		client := getClient()
		s, err := client.Get(isoUrl)
		if err != nil {
			return err
		}
		src = s.Body
	}

	defer src.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}

	defer os.Remove(f.Name())

	if _, err := io.Copy(f, src); err != nil {
		// TODO: display download progress?
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(f.Name(), filepath.Join(dir, file)); err != nil {
		return err
	}

	return nil
}

func (b *B2dUtils) DownloadLatestBoot2Docker() error {
	latestReleaseUrl, err := b.GetLatestBoot2DockerReleaseURL()
	if err != nil {
		return err
	}

	log.Infof("Downloading latest boot2docker release to %s...", b.commonIsoPath)
	if err := b.DownloadISO(b.imgCachePath, b.isoFilename, latestReleaseUrl); err != nil {
		return err
	}

	return nil
}

func (b *B2dUtils) CopyIsoToMachineDir(isoURL, machineName string) error {
	machinesDir := GetMachineDir()
	machineIsoPath := filepath.Join(machinesDir, machineName, b.isoFilename)

	// just in case the cache dir has been manually deleted,
	// check for it and recreate it if it's gone
	if _, err := os.Stat(b.imgCachePath); os.IsNotExist(err) {
		log.Infof("Image cache does not exist, creating it at %s...", b.imgCachePath)
		if err := os.Mkdir(b.imgCachePath, 0700); err != nil {
			return err
		}
	}

	// By default just copy the existing "cached" iso to
	// the machine's directory...
	if isoURL == "" {
		if err := b.copyDefaultIsoToMachine(machineIsoPath); err != nil {
			return err
		}
	} else {
		// But if ISO is specified go get it directly
		log.Infof("Downloading %s from %s...", b.isoFilename, isoURL)
		if err := b.DownloadISO(filepath.Join(machinesDir, machineName), b.isoFilename, isoURL); err != nil {
			return err
		}
	}

	return nil
}

func (b *B2dUtils) copyDefaultIsoToMachine(machineIsoPath string) error {
	if _, err := os.Stat(b.commonIsoPath); os.IsNotExist(err) {
		log.Info("No default boot2docker iso found locally, downloading the latest release...")
		if err := b.DownloadLatestBoot2Docker(); err != nil {
			return err
		}
	}

	if err := CopyFile(b.commonIsoPath, machineIsoPath); err != nil {
		return err
	}

	return nil
}
