package extension

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/log"
)

// ExtensionOptions is the stuct taken as a command line argument.
// This will be a string for where the JSON file is located
type ExtensionOptions struct {
	File string
}

// extensions is used for Detection and registering extensions into a map
var extensions = make(map[string]*RegisteredExtension)

type RegisteredExtension struct {
	New func() Extension
}

func RegisterExtension(name string, e *RegisteredExtension) {
	extensions[name] = e
}

// Used to hold environment variables for /etc/environment
type envs map[string]string

// Used to hold parameters for non-generic extensions (customized installs, etc)
type params map[string]string

// Used to create key:value of files to transfer.
// The interface{} is another map with a key of source, and definition for specific extensions to use
type files map[string]map[string]string

// ExtensionInfo is used in ExtensionInstall. Name is the name of the extension.
// params are the attributes extracted from the JSON file
type ExtensionInfo struct {
	name    string
	version string
	envs    envs
	params  params
	files   files
	copies  files
	run     []string
	validOS []string
}

// ExtensionParams is used in provisionerInfo. This is all the host info needed by the extensions for customized installs
type ExtensionParams struct {
	OsName    string
	OsID      string
	OsVersion string
	Hostname  string
	Ip        string
}

// Extension interface are the actions every extension needs.
// Will need an uninstall and maybe an upgrade later on
type Extension interface {
	//install the extension
	Install(provisioner provision.Provisioner, hostInfo *ExtensionParams, extInfo *ExtensionInfo) error
}

var extInfos []*ExtensionInfo

func ParseExtensionFile(extensionOptions ExtensionOptions) error {
	extensionsToInstall, err := extensionsFile(extensionOptions.File)
	if err != nil {
		return err
	}

	for _, ext := range extensionsToInstall["extensions"].([]interface{}) {
		for k, v := range ext.(map[string]interface{}) {
			//the extensions and it's attributes are saved in a struct
			extInfo := &ExtensionInfo{
				name: k,
			}

			for key, value := range v.(map[string]interface{}) {
				switch key {
				case "version":
					log.Debugf("%s: Parsing version", strings.ToUpper(extInfo.name))
					extInfo.version = value.(string)
					log.Debugf("%s: version=%s", strings.ToUpper(extInfo.name), extInfo.version)
				case "envs":
					log.Debugf("%s: Parsing envs", strings.ToUpper(extInfo.name))
					envs := make(envs)
					for k, v := range value.(map[string]interface{}) {
						envs[k] = v.(string)
					}
					extInfo.envs = envs
					log.Debugf("%s: envs=%v", strings.ToUpper(extInfo.name), extInfo.envs)
				case "params":
					log.Debugf("%s: Parsing params", strings.ToUpper(extInfo.name))
					params := make(params)
					for k, v := range value.(map[string]interface{}) {
						params[k] = v.(string)
					}
					extInfo.params = params
					log.Debugf("%s: params=%v", strings.ToUpper(extInfo.name), extInfo.params)
				case "files":
					log.Debugf("%s: Parsing files", strings.ToUpper(extInfo.name))
					files := make(files)
					for fileskey, filesvalue := range value.(map[string]interface{}) {
						files[fileskey] = make(map[string]string)
						for filekey, filevalue := range filesvalue.(map[string]interface{}) {
							files[fileskey][filekey] = filevalue.(string)
						}
					}
					extInfo.files = files
					log.Debugf("%s: files=%v", strings.ToUpper(extInfo.name), files)
				case "validOS":
					log.Debugf("%s: Parsing validOS", strings.ToUpper(extInfo.name))
					extInfo.validOS = make([]string, 0)
					for _, val := range value.([]interface{}) {
						extInfo.validOS = append(extInfo.validOS, val.(string))
					}
					log.Debugf("%s: validOS=%v", strings.ToUpper(extInfo.name), extInfo.validOS)
				case "run":
					log.Debugf("%s: Parsing run", strings.ToUpper(extInfo.name))
					extInfo.run = make([]string, 0)
					for _, val := range value.([]interface{}) {
						extInfo.run = append(extInfo.run, val.(string))
						log.Debugf("%s: run=%v", strings.ToUpper(extInfo.name), val.(string))
					}
				case "copy":
					copies := make(files)
					var count int
					for fileskey, filesvalue := range value.(map[string]interface{}) {
						keyi := fmt.Sprintf("%s", count)
						copies[keyi] = map[string]string{
							"source":      fileskey,
							"destination": filesvalue.(string),
						}
						log.Debugf("%s: copies[%s]=%v", strings.ToUpper(extInfo.name), keyi, copies[keyi])
						count++
					}
					extInfo.copies = copies
				}
			}
			extInfos = append(extInfos, extInfo)
		}
	}
	return nil
}

// ExtensionInstall function is called from libmachine/host.go in the create function
func ExtensionInstall(provisioner provision.Provisioner) error {

	hostInfo, err := provisonerInfo(provisioner)
	if err != nil {
		return err
	}

	for _, extInfo := range extInfos {
		// FindExtension see if the extension in the JSON file matches a registered extension.
		var extensionFound bool
	FindExtension:
		for extName, extInterface := range extensions {
			switch extInfo.name {
			case extName:
				//create a new interface
				extension := extInterface.New()
				log.Debugf("Found compatible extension: %s", extName)
				//pass everything to the install method and make it happen!
				if err := extension.Install(provisioner, hostInfo, extInfo); err != nil {
					return err
				}
				extensionFound = true
				break FindExtension
			default:
				extensionFound = false
			}
		}
		if extensionFound == false {
			log.Warnf("No compatible extension found for: %s", extInfo.name)
		}
	}
	return nil
}

// extensionsFile is used to parse the extensions JSON file into Go formats
func extensionsFile(filename string) (map[string]interface{}, error) {
	var extI interface{}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("No extensions file specified. Error: %s", err)
	}
	log.Debugf("Parsing information from: %s", filename)
	//determine if file is JSON or YML -- TODO
	//if JSON
	if err := json.Unmarshal([]byte(file), &extI); err != nil {
		return nil, fmt.Errorf("Error parsing JSON. Is it formatted correctly? Error: %s", err)
	}
	// return the extension interface
	return extI.(map[string]interface{}), nil
}

// provisonerInfo Gets all of the host information for the extension to use for installation
func provisonerInfo(provisioner provision.Provisioner) (*ExtensionParams, error) {
	log.Debugf("Gathering Host Information for Extensions")
	os, err := provisioner.GetOsReleaseInfo()
	if err != nil {
		return nil, err
	}

	//may need to look into getting the kernel version if it's necessary
	ip, err := provisioner.GetDriver().GetIP()
	if err != nil {
		return nil, err
	}

	hostname, err := provisioner.Hostname()
	if err != nil {
		return nil, err
	}

	params := ExtensionParams{
		OsName:    os.Name,
		OsID:      os.Id,
		OsVersion: os.Version,
		Hostname:  hostname,
		Ip:        ip,
	}

	return &params, nil
}
