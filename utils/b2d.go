package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	timeout = time.Second * 5
)

func defaultTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, timeout)
}

func getClient() *http.Client {
	transport := http.Transport{
		Dial: defaultTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}

	return &client
}

type B2dUtils struct {
	githubApiBaseUrl string
	githubBaseUrl    string
}

func NewB2dUtils(githubApiBaseUrl, githubBaseUrl string) *B2dUtils {
	defaultBaseApiUrl := "https://api.github.com"
	defaultBaseUrl := "https://github.com"

	if githubApiBaseUrl == "" {
		githubApiBaseUrl = defaultBaseApiUrl
	}

	if githubBaseUrl == "" {
		githubBaseUrl = defaultBaseUrl
	}

	return &B2dUtils{
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
func (b *B2dUtils) DownloadISO(dir, file, url string) error {
	client := getClient()
	rsp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, rsp.Body); err != nil {
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
