package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/state"
)

// Store persists hosts on the filesystem
type Store struct {
	Path           string
	CaCertPath     string
	PrivateKeyPath string
}

func NewStore(rootPath string, caCert string, privateKey string) *Store {
	if rootPath == "" {
		rootPath = filepath.Join(drivers.GetHomeDir(), ".docker", "machines")
	}

	return &Store{Path: rootPath, CaCertPath: caCert, PrivateKeyPath: privateKey}
}

func (s *Store) Create(name string, driverName string, flags drivers.DriverOptions) (*Host, error) {
	exists, err := s.Exists(name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("Host %q already exists", name)
	}

	hostPath := filepath.Join(s.Path, name)

	host, err := NewHost(name, driverName, hostPath, s.CaCertPath, s.PrivateKeyPath)
	if err != nil {
		return host, err
	}
	if flags != nil {
		if err := host.Driver.SetConfigFromFlags(flags); err != nil {
			return host, err
		}
	}

	if err := os.MkdirAll(hostPath, 0700); err != nil {
		return nil, err
	}

	if err := host.SaveConfig(); err != nil {
		return host, err
	}

	if err := host.Create(name); err != nil {
		return host, err
	}

	if err := host.ConfigureAuth(); err != nil {
		return host, err
	}

	return host, nil
}

func (s *Store) Remove(name string, force bool) error {
	active, err := s.GetActive()
	if err != nil {
		return err
	}

	if active != nil && active.Name == name {
		if err := s.RemoveActive(); err != nil {
			return err
		}
	}

	host, err := s.Load(name)
	if err != nil {
		return err
	}
	return host.Remove(force)
}

func (s *Store) List() ([]Host, error) {
	dir, err := ioutil.ReadDir(s.Path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	hosts := []Host{}

	for _, file := range dir {
		if file.IsDir() {
			host, err := s.Load(file.Name())
			if err != nil {
				log.Errorf("error loading host %q: %s", file.Name(), err)
				continue
			}
			hosts = append(hosts, *host)
		}
	}
	return hosts, nil
}

func (s *Store) Exists(name string) (bool, error) {
	_, err := os.Stat(filepath.Join(s.Path, name))
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}
	return false, err
}

func (s *Store) Load(name string) (*Host, error) {
	hostPath := filepath.Join(s.Path, name)
	return LoadHost(name, hostPath)
}

func (s *Store) GetActive() (*Host, error) {
	hostName, err := ioutil.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return s.Load(string(hostName))
}

func (s *Store) IsActive(host *Host) (bool, error) {
	active, err := s.GetActive()
	if err != nil {
		return false, err
	}
	if active == nil {
		return false, nil
	}
	return active.Name == host.Name, nil
}

func (s *Store) SetActive(host *Host) error {
	if err := os.MkdirAll(filepath.Dir(s.activePath()), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(s.activePath(), []byte(host.Name), 0600)
}

func (s *Store) Export(name string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	path := filepath.Join(s.Path, name)
	host, err := LoadHost(name, path)
	if err != nil {
		return buf, err
	}

	driverState, err := host.Driver.GetState()
	if err != nil {
		return buf, fmt.Errorf("Unable to get state")
	}

	if host.Type == "local" && driverState != state.Stopped {
		return buf, fmt.Errorf("Host '%s' should be stopped", name)
	}

	if err := host.Driver.Export(); err != nil {
		return buf, err
	}

	tw := tar.NewWriter(buf)

	walkFunc := func(filePath string, fi os.FileInfo, err error) error {
		re := regexp.MustCompile(fmt.Sprintf("^%s%c", filepath.Join(path, name), os.PathSeparator))

		if len(re.FindString(filePath)) > 0 {
			return nil
		}

		if filepath.Join(path, name) == filePath {
			return nil
		}

		if filePath == path {
			return nil
		}

		if h, err := tar.FileInfoHeader(fi, filePath); err != nil {
			return err
		} else {
			pathName := strings.Replace(filePath, fmt.Sprintf("%s%c", path, os.PathSeparator), "", 1)

			h.Name = pathName
			if err = tw.WriteHeader(h); err != nil {
				return err
			}
		}

		if fi.Mode().IsDir() {
			return nil
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	}

	if err := filepath.Walk(path, walkFunc); err != nil {
		tw.Close()
		return buf, err
	}

	tw.Close()

	return buf, nil
}

func (s *Store) Import(name string, input *os.File) error {
	fi, err := input.Stat()
	if err != nil {
		return err
	}

	if fi.Size() == 0 {
		return fmt.Errorf("no data from stdin")
	}

	path := filepath.Join(s.Path, name)
	if err := os.MkdirAll(path, 0700); err != nil {
		return fmt.Errorf("unable to create machine storage path")
	}

	in := bufio.NewReader(input)
	tr := tar.NewReader(in)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(filepath.Join(path, hdr.Name), 0744); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.Create(filepath.Join(path, hdr.Name))
			if err != nil {
				return fmt.Errorf("Unable to create file '%s'", hdr.Name)
			}

			io.Copy(f, tr)

			f.Close()
		default:
		}
	}

	host, err := LoadHost(name, path)
	if err != nil {
		return err
	}

	if err := host.Driver.Import(name); err != nil {
		return err
	}

	host.SaveConfig()

	return nil
}

func (s *Store) RemoveActive() error {
	return os.Remove(s.activePath())
}

// activePath returns the path to the file that stores the name of the
// active host
func (s *Store) activePath() string {
	return filepath.Join(s.Path, ".active")
}
