package mcnutils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestGetLatestBoot2DockerReleaseUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respText := `[{"tag_name": "0.1"}]`
		w.Write([]byte(respText))
	}))
	defer ts.Close()

	b := NewB2dUtils("/tmp/isos")
	isoUrl, err := b.GetLatestBoot2DockerReleaseURL(ts.URL + "/repos/org/repo/releases")
	if err != nil {
		t.Fatal(err)
	}

	expectedUrl := fmt.Sprintf("%s/org/repo/releases/download/0.1/boot2docker.iso", ts.URL)
	if isoUrl != expectedUrl {
		t.Fatalf("expected url %s; received %s", expectedUrl, isoUrl)
	}
}

func TestDownloadIso(t *testing.T) {
	testData := "test-download"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testData))
	}))
	defer ts.Close()

	filename := "test"

	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}

	b := NewB2dUtils("/tmp/artifacts")
	if err := b.DownloadISO(tmpDir, filename, ts.URL); err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(filepath.Join(tmpDir, filename))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testData {
		t.Fatalf("expected data \"%s\"; received \"%s\"", testData, string(data))
	}
}

func TestGetReleasesRequestNoToken(t *testing.T) {
	GithubApiToken = ""
	b2d := NewB2dUtils("/tmp/store")
	req, err := b2d.getReleasesRequest("http://some.github.api")
	if err != nil {
		t.Fatal("Expected err to be nil, got ", err)
	}

	if req.Header.Get("Authorization") != "" {
		t.Fatal("Expected not to get an 'Authorization' header, but got one: ", req.Header.Get("Authorization"))
	}
}

func TestGetReleasesRequest(t *testing.T) {
	expectedToken := "CATBUG"
	GithubApiToken = expectedToken
	b2d := NewB2dUtils("/tmp/store")

	req, err := b2d.getReleasesRequest("http://some.github.api")
	if err != nil {
		t.Fatal("Expected err to be nil, got ", err)
	}

	if req.Header.Get("Authorization") != fmt.Sprintf("token %s", expectedToken) {
		t.Fatal("Header was not set as expected: ", req.Header.Get("Authorization"))
	}
}
