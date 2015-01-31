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

func GetBoot2DockerISO(skipUpdate bool, boot2DockerURL string, storePath string) error {
	if !skipUpdate {

		if boot2DockerURL != "" {
			isoURL := boot2DockerURL
			log.Infof("Downloading boot2docker.iso from %s...", isoURL)
			if err := DownloadISO(storePath, "boot2docker.iso", isoURL); err != nil {
				return err
			}
		} else {
			// todo: check latest release URL, download if it's new
			// until then always use "latest"
			isoURL, err := GetLatestBoot2DockerReleaseURL()
			if err != nil {
				return err
			}

			// todo: use real constant for .docker
			rootPath := filepath.Join(GetHomeDir(), ".docker")
			imgPath := filepath.Join(rootPath, "images")
			commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")
			if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
				log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)

				// just in case boot2docker.iso has been manually deleted
				if _, err := os.Stat(imgPath); os.IsNotExist(err) {
					if err := os.Mkdir(imgPath, 0700); err != nil {
						return err
					}
				}

				if err := DownloadISO(imgPath, "boot2docker.iso", isoURL); err != nil {
					return err
				}
			}

			isoDest := filepath.Join(storePath, "boot2docker.iso")
			if err := cpIso(commonIsoPath, isoDest); err != nil {
				return err
			}
		}
	} else {

		rootPath := filepath.Join(GetHomeDir(), ".docker")
		imgPath := filepath.Join(rootPath, "images")
		commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")

		isoDest := filepath.Join(storePath, "boot2docker.iso")
		if err := cpIso(commonIsoPath, isoDest); err != nil {
			return err
		}

	}
	return nil
}

func cpIso(src, dest string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dest, buf, 0600); err != nil {
		return err
	}
	return nil
}


