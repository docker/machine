package image

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

var (
	ErrNoDownloadUrl = errors.New("No download URL found")
)

type Metadata map[string]string

type DownloadStrategy interface {
	DownloadUrl() (string, Metadata error)
}

type Boot2DockerDownloadStrategy struct {
	Client *http.Client
}

func (s *Boot2DockerDownloadStrategy) DownloadUrl() (string, Metadata, error) {
	api := "https://api.github.com"
	web := "https://github.com"

	if len(os.Getenv("MACHINE_B2D_GITHUB_ENTERPRISE_API")) != 0 {
		api = os.Getenv("MACHINE_B2D_GITHUB_ENTERPRISE_API")
	}

	if len(os.Getenv("MACHINE_B2D_GITHUB_ENTERPRISE_WEB")) != 0 {
		api = os.Getenv("MACHINE_B2D_GITHUB_ENTERPRISE_WEB")
	}

	meta := Metadata{}

	resp, err := s.Client.Get(fmt.Sprintf("%s/repos/boot2docker/boot2docker/releases", api))
	if err != nil {
		return "", meta, err
	}

	defer resp.Body.Close()

	var t []struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
		return "", err
	}

	if len(t) == 0 {
		return "", ErrNoDownloadUrl
	}

	tag := t[0].TagName

	meta["version"] = tag

	iso := fmt.Sprintf(
		"%s/boot2docker/boot2docker/releases/download/%s/boot2docker.iso",
		web,
		tag,
	)

	return iso, meta, nil
}

type LocalDownloadStrategy struct {
	Path string
	Stat func(path string) (os.FileInfo, err)
}

func (s *LocalDownloadStrategy) DownloadUrl() (string, Metadata, error) {
	if s.Stat == nil {
		s.Stat = os.Stat
	}

	if _, err := s.Stat(s.Path); err != nil {
		return "", Metadata{}, ErrNoDownloadUrl
	}

	return s.Path, Metadata{}, nil
}
