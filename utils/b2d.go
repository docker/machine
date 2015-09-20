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

	"github.com/docker/machine/log"
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
	githubAPIBaseURL string
	githubBaseURL    string
}

func NewB2dUtils(githubAPIBaseURL, githubBaseURL string) *B2dUtils {
	defaultBaseAPIURL := "https://api.github.com"
	defaultBaseURL := "https://github.com"
	imgCachePath := GetMachineCacheDir()
	isoFilename := "boot2docker.iso"

	if githubAPIBaseURL == "" {
		githubAPIBaseURL = defaultBaseAPIURL
	}

	if githubBaseURL == "" {
		githubBaseURL = defaultBaseURL
	}

	return &B2dUtils{
		isoFilename:      isoFilename,
		imgCachePath:     GetMachineCacheDir(),
		commonIsoPath:    filepath.Join(imgCachePath, isoFilename),
		githubAPIBaseURL: githubAPIBaseURL,
		githubBaseURL:    githubBaseURL,
	}
}

// GetLatestBoot2DockerReleaseURL gets the latest boot2docker release
// url / tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of
// boot2docker since the GitHub API has a pretty low rate limit on API
// requests
func (b *B2dUtils) GetLatestBoot2DockerReleaseURL() (string, error) {
	client := getClient()
	apiURL := fmt.Sprintf("%s/repos/boot2docker/boot2docker/releases", b.githubAPIBaseURL)
	rsp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	var t []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
		return "", fmt.Errorf("Error demarshaling the Github API response: %s\nYou may be getting rate limited by Github.", err)
	}
	if len(t) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	tag := t[0].TagName
	isoURL := fmt.Sprintf("%s/boot2docker/boot2docker/releases/download/%s/boot2docker.iso", b.githubBaseURL, tag)
	return isoURL, nil
}

func removeFileIfExists(name string) error {
	if _, err := os.Stat(name); err == nil {
		if err := os.Remove(name); err != nil {
			log.Fatalf("Error removing temporary download file: %s", err)
		}
	}
	return nil
}

// DownloadISO gets boot2docker ISO image from isoURL and save it at dir/file.
func (b *B2dUtils) DownloadISO(dir, file, isoURL string) error {
	u, err := url.Parse(isoURL)
	var src io.ReadCloser
	if u.Scheme == "file" || u.Scheme == "" {
		s, err := os.Open(u.Path)
		if err != nil {
			return err
		}
		src = s
	} else {
		client := getClient()
		s, err := client.Get(isoURL)
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

	defer func() {
		if err := removeFileIfExists(f.Name()); err != nil {
			log.Fatalf("Error removing file: %s", err)
		}
	}()

	if _, err := io.Copy(f, src); err != nil {
		// TODO: display download progress?
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Dest is the final path of the boot2docker.iso file.
	dest := filepath.Join(dir, file)

	// Windows can't rename in place, so remove the old file before
	// renaming the temporary downloaded file.
	if err := removeFileIfExists(dest); err != nil {
		return err
	}

	if err := os.Rename(f.Name(), dest); err != nil {
		return err
	}

	return nil
}

// DownloadLatestBoot2Docker gets the latest url, and calls
// DownloadISOFromURL to fetch it
func (b *B2dUtils) DownloadLatestBoot2Docker() error {
	latestReleaseURL, err := b.GetLatestBoot2DockerReleaseURL()
	if err != nil {
		return err
	}

	return b.DownloadISOFromURL(latestReleaseURL)
}

func (b *B2dUtils) DownloadISOFromURL(latestReleaseURL string) error {
	log.Infof("Downloading %s to %s...", latestReleaseURL, b.commonIsoPath)
	if err := b.DownloadISO(b.imgCachePath, b.isoFilename, latestReleaseURL); err != nil {
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
