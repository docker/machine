package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

func DownloadUpdateB2D() (string, error) {
	//todo make this func update the already downloaded b2d if there is a newer version
	isoURL, err := GetLatestBoot2DockerReleaseURL()
	if err != nil {
		return "", err
	}

	imgPath := filepath.Join(GetDockerDir(), "images")
	commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")
	if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
		log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)

		// just in case boot2docker.iso has been manually deleted
		if _, err := os.Stat(imgPath); os.IsNotExist(err) {
			if err := os.Mkdir(imgPath, 0700); err != nil {
				return "", err
			}
		}

		if err := DownloadISO(imgPath, "boot2docker.iso", isoURL); err != nil {
			return "", err
		}
	}
	return commonIsoPath, nil
}

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
func DownloadISO(dir, file, url string) error {
	rsp, err := http.Get(url)
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
