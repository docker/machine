package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

// Get the latest boot2docker release tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func GetLatestBoot2DockerReleaseURL() (string, error) {
	rsp, err := http.Get("https://api.github.com/repos/boot2docker/boot2docker/releases")
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
	url := fmt.Sprintf("https://github.com/boot2docker/boot2docker/releases/download/%s/boot2docker.iso", tag)
	return url, nil
}

// Download boot2docker ISO image for the given tag and save it at dest.
func DownloadISO(dir, file, u string) error {

	uri, err := url.Parse(u)
	if err != nil {
		return err
	}

	if uri.Scheme == "file" || uri.Scheme == "" {
		src, err := os.Open(uri.Path)
		if err != nil {
			return err
		}
		defer src.Close()
		return cp(dir, file, src)

	} else {

		rsp, err := http.Get(u)
		if err != nil {
			return err
		}
		defer rsp.Body.Close()
		return cp(dir, file, rsp.Body)
	}
}

func cp(dir, file string, src io.Reader) error {

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	// TODO: display download progress?
	if _, err := io.Copy(f, src); err != nil {
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
