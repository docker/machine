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

	b := NewB2dUtils(ts.URL, ts.URL, "/tmp/isos")
	isoUrl, err := b.GetLatestBoot2DockerReleaseURL()
	if err != nil {
		t.Fatal(err)
	}

	expectedUrl := fmt.Sprintf("%s/boot2docker/boot2docker/releases/download/0.1/boot2docker.iso", ts.URL)
	if isoUrl != expectedUrl {
		t.Fatalf("expected url %s; received %s", isoUrl, expectedUrl)
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

	b := NewB2dUtils(ts.URL, ts.URL, "/tmp/artifacts")
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
