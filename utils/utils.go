package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/docker/machine/log"
)

func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

func GetBaseDir() string {
	baseDir := os.Getenv("MACHINE_STORAGE_PATH")
	if baseDir == "" {
		baseDir = filepath.Join(GetHomeDir(), ".docker", "machine")
	}
	return baseDir
}

func GetDockerDir() string {
	return filepath.Join(GetHomeDir(), ".docker")
}

func GetMachineDir() string {
	return filepath.Join(GetBaseDir(), "machines")
}

func GetMachineCertDir() string {
	return filepath.Join(GetBaseDir(), "certs")
}

func GetMachineCacheDir() string {
	return filepath.Join(GetBaseDir(), "cache")
}

func GetUsername() string {
	u := "unknown"
	osUser := ""

	switch runtime.GOOS {
	case "darwin", "linux":
		osUser = os.Getenv("USER")
	case "windows":
		osUser = os.Getenv("USERNAME")
	}

	if osUser != "" {
		u = osUser
	}

	return u
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.Chmod(dst, fi.Mode()); err != nil {
		return err
	}

	return nil
}

func WaitForSpecific(f func() bool, maxAttempts int, waitInterval time.Duration) error {
	for i := 0; i < maxAttempts; i++ {
		if f() {
			return nil
		}
		time.Sleep(waitInterval)
	}
	return fmt.Errorf("Maximum number of retries (%d) exceeded", maxAttempts)
}

func WaitFor(f func() bool) error {
	return WaitForSpecific(f, 60, 3*time.Second)
}

func WaitForDocker(ip string, daemonPort int) error {
	return WaitFor(func() bool {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, daemonPort))
		if err != nil {
			log.Debugf("Daemon not responding yet: %s", err)
			return false
		}
		conn.Close()
		return true
	})
}

func DumpVal(vals ...interface{}) {
	for _, val := range vals {
		prettyJSON, err := json.MarshalIndent(val, "", "    ")
		if err != nil {
			log.Fatal(err)
		}
		log.Debug(string(prettyJSON))
	}
}

// Following two functions are from github.com/docker/docker/utils module. It
// was way overkill to include the whole module, so we just have these bits
// that we're using here.
func TruncateID(id string) string {
	shortLen := 12
	if len(id) < shortLen {
		shortLen = len(id)
	}
	return id[:shortLen]
}

// GenerateRandomID returns an unique id
func GenerateRandomID() string {
	for {
		id := make([]byte, 32)
		if _, err := io.ReadFull(rand.Reader, id); err != nil {
			panic(err) // This shouldn't happen
		}
		value := hex.EncodeToString(id)
		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numberic and causes issues when
		// used as a hostname. ref #3869
		if _, err := strconv.ParseInt(TruncateID(value), 10, 64); err == nil {
			continue
		}
		return value
	}
}
