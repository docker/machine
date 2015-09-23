package persisttest

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/docker/machine/libmachine/hosttest"
	"github.com/docker/machine/libmachine/persist"
)

var (
	TestStoreDir = ""
)

func Cleanup() error {
	return os.RemoveAll(TestStoreDir)
}

func GetDefaultTestStore() (persist.Filestore, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	TestStoreDir = tmpDir

	return persist.Filestore{
		Path:             tmpDir,
		CaCertPath:       hosttest.HostTestCaCert,
		CaPrivateKeyPath: hosttest.HostTestPrivateKey,
	}, nil
}
