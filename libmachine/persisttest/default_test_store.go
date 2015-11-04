package persisttest

import "os"

var TestStoreDir = ""

func Cleanup() error {
	return os.RemoveAll(TestStoreDir)
}
