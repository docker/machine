package mcnutils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"bytes"

	"github.com/stretchr/testify/assert"
)

func TestGetLatestBoot2DockerReleaseUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respText := `[{"tag_name": "0.1"}]`
		w.Write([]byte(respText))
	}))
	defer ts.Close()

	b := NewB2dUtils("/tmp/isos")
	isoURL, err := b.GetLatestBoot2DockerReleaseURL(ts.URL + "/repos/org/repo/releases")

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s/org/repo/releases/download/0.1/boot2docker.iso", ts.URL), isoURL)
}

func TestDownloadIso(t *testing.T) {
	testData := "test-download"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(testData))
	}))
	defer ts.Close()

	filename := "test"

	tmpDir, err := ioutil.TempDir("", "machine-test-")

	assert.NoError(t, err)

	b := NewB2dUtils("/tmp/artifacts")
	err = b.DownloadISO(tmpDir, filename, ts.URL)

	assert.NoError(t, err)

	data, err := ioutil.ReadFile(filepath.Join(tmpDir, filename))

	assert.NoError(t, err)
	assert.Equal(t, testData, string(data))
}

func TestGetReleasesRequestNoToken(t *testing.T) {
	GithubAPIToken = ""

	b2d := NewB2dUtils("/tmp/store")
	req, err := b2d.getReleasesRequest("http://some.github.api")

	assert.NoError(t, err)
	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestGetReleasesRequest(t *testing.T) {
	expectedToken := "CATBUG"
	GithubAPIToken = expectedToken

	b2d := NewB2dUtils("/tmp/store")
	req, err := b2d.getReleasesRequest("http://some.github.api")

	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("token %s", expectedToken), req.Header.Get("Authorization"))
}

type MockReadCloser struct {
	blockLengths []int
	currentBlock int
}

func (r *MockReadCloser) Read(p []byte) (n int, err error) {
	n = r.blockLengths[r.currentBlock]
	r.currentBlock++
	return
}

func (r *MockReadCloser) Close() error {
	return nil
}

func TestReaderWithProgress(t *testing.T) {
	readCloser := MockReadCloser{blockLengths: []int{5, 45, 50}}
	output := new(bytes.Buffer)
	buffer := make([]byte, 100)

	readerWithProgress := ReaderWithProgress{
		ReadCloser:     &readCloser,
		out:            output,
		expectedLength: 100,
	}

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%..", output.String())

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%....10%....20%....30%....40%....50%", output.String())

	readerWithProgress.Read(buffer)
	assert.Equal(t, "0%....10%....20%....30%....40%....50%....60%....70%....80%....90%....100%", output.String())

	readerWithProgress.Close()
	assert.Equal(t, "0%....10%....20%....30%....40%....50%....60%....70%....80%....90%....100%\n", output.String())
}
