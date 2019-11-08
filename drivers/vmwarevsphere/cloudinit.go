package vmwarevsphere

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/docker/machine/libmachine/log"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"gopkg.in/yaml.v2"
)

const (
	isoName = "user-data.iso"
	isoDir  = "cloudinit"
)

func (d *Driver) cloudInit(vm *object.VirtualMachine) error {
	if d.CloudInit != "" {
		return d.cloudInitGuestInfo(vm)
	}

	if d.CloudConfig == "" {
		if err := d.createCloudInitIso(); err != nil {
			return err
		}

		ds, err := d.getVmDatastore(vm)
		if err != nil {
			return err
		}

		err = d.uploadCloudInitIso(vm, d.datacenter, ds)
		if err != nil {
			return err
		}

		err = d.mountCloudInitIso(vm, d.datacenter, ds)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) cloudInitGuestInfo(vm *object.VirtualMachine) error {
	var opts []types.BaseOptionValue
	if d.CloudInit != "" {
		if _, err := url.ParseRequestURI(d.CloudInit); err == nil {
			log.Infof("setting guestinfo.cloud-init.data.url to %s\n", d.CloudInit)
			opts = append(opts, &types.OptionValue{
				Key:   "guestinfo.cloud-init.config.url",
				Value: d.CloudInit,
			})
		} else {
			if _, err := os.Stat(d.CloudInit); err == nil {
				if value, err := ioutil.ReadFile(d.CloudInit); err == nil {
					log.Infof("setting guestinfo.cloud-init.data to encoded content of %s\n", d.CloudInit)
					encoded := base64.StdEncoding.EncodeToString(value)
					opts = append(opts, &types.OptionValue{
						Key:   "guestinfo.cloud-init.config.data",
						Value: encoded,
					})
					opts = append(opts, &types.OptionValue{
						Key:   "guestinfo.cloud-init.data.encoding",
						Value: "base64",
					})
				}
			}
		}
	}

	return d.applyOpts(vm, opts)
}

func (d *Driver) uploadCloudInitIso(vm *object.VirtualMachine, dc *object.Datacenter, ds *object.Datastore) error {
	log.Infof("Uploading cloud-init.iso")
	path, err := d.getVmFolder(vm)
	if err != nil {
		return err
	}

	dsurl, err := ds.URL(d.getCtx(), dc, filepath.Join(path, isoName))
	if err != nil {
		return err
	}

	p := soap.DefaultUpload
	c, err := d.getSoapClient()
	if err != nil {
		return err
	}

	if err = c.Client.UploadFile(d.getCtx(), d.ResolveStorePath(filepath.Join(isoDir, isoName)), dsurl, &p); err != nil {
		return err
	}

	return nil
}

func (d *Driver) removeCloudInitIso(vm *object.VirtualMachine, dc *object.Datacenter, ds *object.Datastore) error {
	log.Infof("Removing cloud-init.iso")
	c, err := d.getSoapClient()
	if err != nil {
		return err
	}

	path, err := d.getVmFolder(vm)
	if err != nil {
		return err
	}

	m := object.NewFileManager(c.Client)
	task, err := m.DeleteDatastoreFile(d.getCtx(), ds.Path(filepath.Join(path, isoName)), dc)
	if err != nil {
		return err
	}

	if err = task.Wait(d.getCtx()); err != nil {
		if types.IsFileNotFound(err) {
			// already deleted, ignore error
			return nil
		}

		return err
	}

	return nil
}

func (d *Driver) createCloudInitIso() error {
	log.Infof("Creating cloud-init.iso")
	//d.CloudConfig stat'ed and loaded in flag load.
	sshkey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	userdatacontent, err := addSSHUserToYaml(d.CloudConfig, d.SSHUser, string(sshkey))
	if err != nil {
		return err
	}

	perm := os.FileMode(0700)
	isoDir := d.ResolveStorePath(isoDir)
	dataDir := filepath.Join(isoDir, "data")
	userdata := filepath.Join(dataDir, "user-data")
	metadata := filepath.Join(dataDir, "meta-data")

	err = os.MkdirAll(dataDir, perm)
	if err != nil {
		return err
	}

	writeYaml := fmt.Sprintf("#cloud-config\n%s", userdatacontent)
	if err = ioutil.WriteFile(userdata, []byte(writeYaml), perm); err != nil {
		return err
	}

	md := []byte(fmt.Sprintf("#local-hostname: %s\n", d.MachineName))
	if err = ioutil.WriteFile(metadata, md, perm); err != nil {
		return err
	}

	//making iso
	blocksize := int64(2048)
	diskImg := filepath.Join(isoDir, isoName)
	iso, err := os.OpenFile(diskImg, os.O_CREATE|os.O_RDWR, perm)
	if err != nil {
		return err
	}
	defer iso.Close()

	fs, err := iso9660.Create(iso, 0, 0, blocksize)
	if err != nil {
		return err
	}

	err = fs.Mkdir("/")
	if err != nil {
		return err
	}

	for filename, filepath := range map[string]string{"user-data": userdata, "meta-data": metadata} {
		f, err := ioutil.ReadFile(filepath) // just pass the file name
		if err != nil {
			return err
		}

		rw, err := fs.OpenFile("/"+filename, os.O_CREATE|os.O_RDWR)
		if err != nil {
			return err
		}

		_, err = rw.Write(f)
		if err != nil {
			return err
		}
	}

	return fs.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		VolumeIdentifier: "cidata",
	})
}

func (d *Driver) mountCloudInitIso(vm *object.VirtualMachine, dc *object.Datacenter, dss *object.Datastore) error {
	log.Debugf("Mounting cloudinit %s", isoName)
	devices, err := vm.Device(d.getCtx())
	if err != nil {
		return err
	}

	ide, err := devices.FindIDEController("")
	if err != nil {
		return err
	}

	var add []types.BaseVirtualDevice
	cdrom, err := devices.CreateCdrom(ide)
	if err != nil {
		return err
	}

	path, err := d.getVmFolder(vm)
	if err != nil {
		return err
	}

	add = append(add, devices.InsertIso(cdrom, dss.Path(filepath.Join(path, isoName))))
	return vm.AddDevice(d.getCtx(), add...)
}

func addSSHUserToYaml(yamlcontent, user, sshkey string) (string, error) {
	cf := make(map[interface{}]interface{})
	if err := yaml.Unmarshal([]byte(yamlcontent), &cf); err != nil {
		return "", err
	}

	newUser := map[interface{}]interface{}{
		"name":        user,
		"sudo":        "ALL=(ALL) NOPASSWD:ALL",
		"lock_passwd": "true",
		"ssh_authorized_keys": []string{
			sshkey,
		},
	}

	if val, ok := cf["users"]; ok {
		u := val.([]interface{})
		u = append(u, newUser)
		cf["users"] = u
	} else {
		users := make([]interface{}, 1)
		users[0] = newUser
		cf["users"] = users
	}

	yaml, err := yaml.Marshal(cf)
	if err != nil {
		return "", err
	}

	return string(yaml), nil
}
