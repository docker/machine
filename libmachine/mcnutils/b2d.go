package mcnutils

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
	"regexp"

	"github.com/docker/machine/libmachine/log"
)

var (
	GithubAPIToken string
)

func defaultTimeout(network, addr string) (net.Conn, error) {
	return net.Dial(network, addr)
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
	storePath        string
	isoFilename      string
	commonIsoPath    string
	imgCachePath     string
	githubAPIBaseURL string
	githubBaseURL    string
}

func NewB2dUtils(storePath string) *B2dUtils {
	imgCachePath := filepath.Join(storePath, "cache")
	isoFilename := "boot2docker.iso"

	return &B2dUtils{
		storePath:     storePath,
		isoFilename:   isoFilename,
		imgCachePath:  imgCachePath,
		commonIsoPath: filepath.Join(imgCachePath, isoFilename),
	}
}

func (b *B2dUtils) getReleasesRequest(apiURL string) (*http.Request, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if GithubAPIToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", GithubAPIToken))
	}

	return req, nil
}

// GetLatestBoot2DockerReleaseURL gets the latest boot2docker release tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func (b *B2dUtils) GetLatestBoot2DockerReleaseURL(apiURL string) (string, error) {
	if apiURL == "" {
		apiURL = "https://api.github.com/repos/boot2docker/boot2docker/releases"
	}
	isoURL := ""
	// match github (enterprise) release urls:
	// https://api.github.com/repos/../../releases or
	// https://some.github.enterprise/api/v3/repos/../../releases
	re := regexp.MustCompile("(https?)://([^/]+)(/api/v3)?/repos/([^/]+)/([^/]+)/releases")
	if matches := re.FindStringSubmatch(apiURL); len(matches) == 6 {
		scheme := matches[1]
		host := matches[2]
		org := matches[4]
		repo := matches[5]
		if host == "api.github.com" {
			host = "github.com"
		}
		client := getClient()
		req, err := b.getReleasesRequest(apiURL)
		if err != nil {
			return "", err
		}
		rsp, err := client.Do(req)
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
		log.Infof("Latest release for %s/%s/%s is %s", host, org, repo, tag)
		isoURL = fmt.Sprintf("%s://%s/%s/%s/releases/download/%s/boot2docker.iso", scheme, host, org, repo, tag)
	} else {
		//does not match a github releases api url
		isoURL = apiURL
	}

	return isoURL, nil
}

func removeFileIfExists(name string) error {
	if _, err := os.Stat(name); err == nil {
		if err := os.Remove(name); err != nil {
			return fmt.Errorf("Error removing temporary download file: %s", err)
		}
	}
	return nil
}

// DownloadISO downloads boot2docker ISO image for the given tag and save it at dest.
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

		src = &ReaderWithProgress{
			ReadCloser:     s.Body,
			out:            os.Stdout,
			expectedLength: s.ContentLength,
		}
	}

	defer src.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}

	defer func() {
		if err := removeFileIfExists(f.Name()); err != nil {
			log.Warnf("Error removing file: %s", err)
		}
	}()

	if _, err := io.Copy(f, src); err != nil {
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

type ReaderWithProgress struct {
	io.ReadCloser
	out                io.Writer
	bytesTransferred   int64
	expectedLength     int64
	nextPercentToPrint int64
}

func (r *ReaderWithProgress) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)

	if n > 0 {
		r.bytesTransferred += int64(n)
		percentage := r.bytesTransferred * 100 / r.expectedLength

		for percentage >= r.nextPercentToPrint {
			if r.nextPercentToPrint%10 == 0 {
				fmt.Fprintf(r.out, "%d%%", r.nextPercentToPrint)
			} else if r.nextPercentToPrint%2 == 0 {
				fmt.Fprint(r.out, ".")
			}
			r.nextPercentToPrint += 2
		}
	}

	return n, err
}

func (r *ReaderWithProgress) Close() error {
	fmt.Fprintln(r.out)
	return r.ReadCloser.Close()
}

func (b *B2dUtils) DownloadLatestBoot2Docker(apiURL string) error {
	latestReleaseURL, err := b.GetLatestBoot2DockerReleaseURL(apiURL)
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
	// TODO: This is a bit off-color.
	machineDir := filepath.Join(b.storePath, "machines", machineName)
	machineIsoPath := filepath.Join(machineDir, b.isoFilename)

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
		//if ISO is specified, check if it matches a github releases url or fallback
		//to a direct download
		if downloadURL, err := b.GetLatestBoot2DockerReleaseURL(isoURL); err == nil {
			log.Infof("Downloading %s from %s...", b.isoFilename, downloadURL)
			if err := b.DownloadISO(machineDir, b.isoFilename, downloadURL); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (b *B2dUtils) copyDefaultIsoToMachine(machineIsoPath string) error {
	if _, err := os.Stat(b.commonIsoPath); os.IsNotExist(err) {
		log.Info("No default boot2docker iso found locally, downloading the latest release...")
		if err := b.DownloadLatestBoot2Docker(""); err != nil {
			return err
		}
	}

	if err := CopyFile(b.commonIsoPath, machineIsoPath); err != nil {
		return err
	}

	return nil
}
