package mcnutils

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	testStr := "test-machine"

	srcFile, err := ioutil.TempFile("", "machine-test-")
	if err != nil {
		t.Fatal(err)
	}
	srcFi, err := srcFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	srcFile.Write([]byte(testStr))
	srcFile.Close()

	srcFilePath := filepath.Join(os.TempDir(), srcFi.Name())

	destFile, err := ioutil.TempFile("", "machine-copy-test-")
	if err != nil {
		t.Fatal(err)
	}

	destFi, err := destFile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	destFile.Close()

	destFilePath := filepath.Join(os.TempDir(), destFi.Name())

	if err := CopyFile(srcFilePath, destFilePath); err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile(destFilePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testStr {
		t.Fatalf("expected data \"%s\"; received \"%s\"", testStr, string(data))
	}
}

func TestGetUsername(t *testing.T) {
	username := GetUsername()

	if username == "unknown" {
		t.Fatalf("Couldn't find username. Got %s", username)
	}
}

func TestGenerateRandomID(t *testing.T) {
	id := GenerateRandomID()

	if len(id) != 64 {
		t.Fatalf("Id returned is incorrect: %s", id)
	}
}

func TestShortenId(t *testing.T) {
	id := GenerateRandomID()
	truncID := TruncateID(id)
	if len(truncID) != 12 {
		t.Fatalf("Id returned is incorrect: truncate on %s returned %s", id, truncID)
	}
}

func TestShortenIdEmpty(t *testing.T) {
	id := ""
	truncID := TruncateID(id)
	if len(truncID) > len(id) {
		t.Fatalf("Id returned is incorrect: truncate on %s returned %s", id, truncID)
	}
}

func TestShortenIdInvalid(t *testing.T) {
	id := "1234"
	truncID := TruncateID(id)
	if len(truncID) != len(id) {
		t.Fatalf("Id returned is incorrect: truncate on %s returned %s", id, truncID)
	}
}
