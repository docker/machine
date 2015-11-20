package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnerror"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/swarm"
)

type machineConfig struct {
	Name           string                 `yaml:"name"`
	Driver         string                 `yaml:"driver"`
	StoragePath    string                 `yaml:"storage_path"`
	CaCertPath     string                 `yaml:"tls_ca_cert"`
	CaKeyPath      string                 `yaml:"tls_ca_key"`
	ClientCertPath string                 `yaml:"tls_client_cert"`
	ClientKeyPath  string                 `yaml:"tls_client_key"`
	GithubAPIToken string                 `yaml:"github_api_token"`
	NativeSSH      bool                   `yaml:"native_ssh"`
	DriverOptions  map[string]interface{} `yaml:"driveroptions"`
	EngineOptions  engine.Options         `yaml:"engineoptions"`
	SwarmOptions   swarm.Options          `yaml:"swarmoptions"`
}

var (
	sharedApplyFlags = []cli.Flag{
		cli.StringFlag{
			Name: "config, c",
			Usage: fmt.Sprintf(
				"Machine declarative config file",
			),
			Value: "docker-machine.yml",
		},
	}
)

func cmdApply(c CommandLine, api libmachine.API) error {

	configFile := c.String("config")

	if configFile == "" {
		c.ShowHelp()
		return nil
	}

	//Read the config file.
	file, err := ioutil.ReadFile(configFile)
	if err != nil {
		println("File does not exist:", err.Error())
		return err
	}

	var mc []machineConfig
	err = yaml.Unmarshal(file, &mc)
	if err != nil {
		fmt.Println("Error while parsing the config file", err)
		return nil
	}

	for _, node := range mc {
		driverName := node.Driver
		if driverName == "" {
			return fmt.Errorf("Driver missing for %s node\n", node.Name)
		}

		// Set default storage-path, if not provided in config file.
		storagePath := node.StoragePath
		if storagePath == "" {
			storagePath = mcndirs.GetBaseDir()
		}

		certInfo := getCertPathInfoFromConfig(node)

		swarmOpt := getSwarmOptions(node.SwarmOptions)

		validName := host.ValidateHostName(node.Name)
		if !validName {
			return fmt.Errorf("Error creating machine: %s", mcnerror.ErrInvalidHostname)
		}

		bareDriverData, err := json.Marshal(&drivers.BaseDriver{
			MachineName: node.Name,
			StorePath:   storagePath,
		})
		if err != nil {
			return fmt.Errorf("Error attempting to marshal bare driver data: %s", err)
		}

		driver, err := api.NewPluginDriver(driverName, bareDriverData)
		if err != nil {
			return fmt.Errorf("Error loading driver %s : %s", driverName, err)
		}

		h, err := api.NewHost(driver)
		if err != nil {
			return fmt.Errorf("Error getting new host: %s", err)
		}

		h.HostOptions = &host.Options{
			AuthOptions: &auth.Options{
				CertDir:          mcndirs.GetMachineCertDir(),
				CaCertPath:       certInfo.CaCertPath,
				CaPrivateKeyPath: certInfo.CaPrivateKeyPath,
				ClientCertPath:   certInfo.ClientCertPath,
				ClientKeyPath:    certInfo.ClientKeyPath,
				ServerCertPath:   filepath.Join(mcndirs.GetMachineDir(), node.Name, "server.pem"),
				ServerKeyPath:    filepath.Join(mcndirs.GetMachineDir(), node.Name, "server-key.pem"),
				StorePath:        filepath.Join(mcndirs.GetMachineDir(), node.Name),
			},
			EngineOptions: &engine.Options{
				ArbitraryFlags:   node.EngineOptions.ArbitraryFlags,
				Env:              node.EngineOptions.Env,
				InsecureRegistry: node.EngineOptions.InsecureRegistry,
				Labels:           node.EngineOptions.Labels,
				RegistryMirror:   node.EngineOptions.RegistryMirror,
				StorageDriver:    node.EngineOptions.StorageDriver,
				TLSVerify:        true,
				InstallURL:       node.EngineOptions.InstallURL,
			},
			SwarmOptions: &swarm.Options{
				IsSwarm:        swarmOpt.IsSwarm,
				Image:          swarmOpt.Image,
				Master:         swarmOpt.Master,
				Discovery:      swarmOpt.Discovery,
				Address:        swarmOpt.Address,
				Host:           swarmOpt.Host,
				Strategy:       swarmOpt.Strategy,
				ArbitraryFlags: swarmOpt.ArbitraryFlags,
			},
		}

		exists, err := api.Exists(h.Name)
		if err != nil {
			return fmt.Errorf("Error checking if host exists: %s", err)
		}
		if exists {
			log.Infof("Machine Exists %s, Trying to start", h.Name)
			mncState, err := h.Driver.GetState()
			if err != nil {
				return fmt.Errorf("Error checking host state: %s", err)
			}
			if mncState != state.Stopped {
				log.Infof("Machine %s is in state %s. Skipping..", h.Name, mncState.String())
				continue
			}
			err = startMachine(driver, node, h, api)
			if err != nil {
				log.Info("Error while starting machine")
				return err
			}
		} else {
			// Create the machine
			if err = createMachine(driver, node, h, api); err != nil {
				log.Info("Error while machine creation: %s", err)
				return err
			}
		}

	}

	return nil

}

// Returns driver options from machineConfig or sets to defaults.
func getDriverOptions(mc machineConfig, mcnflags []mcnflag.Flag) drivers.DriverOptions {
	driverOpts := rpcdriver.RPCFlags{
		Values: make(map[string]interface{}),
	}

	for _, f := range mcnflags {
		driverOpts.Values[f.String()] = f.Default()

		if f.Default() == nil {
			driverOpts.Values[f.String()] = false
		}
	}

	for name, value := range mc.DriverOptions {
		drvOption := mc.Driver + "_" + name
		switch v := value.(type) {
		case int32, int64:
			// typecast to int required, else value become 0 in RPC call.
			driverOpts.Values[yaml2cmd(drvOption)] = int(v.(int64))
		case bool:
			driverOpts.Values[yaml2cmd(drvOption)] = bool(v)
		case string:
			driverOpts.Values[yaml2cmd(drvOption)] = string(v)
		default:
			driverOpts.Values[yaml2cmd(drvOption)] = v
		}
	}

	return driverOpts

}

// Returns the cert paths
// from machineConfig or default
func getCertPathInfoFromConfig(mc machineConfig) cert.PathInfo {
	caCertPath := mc.CaCertPath
	caKeyPath := mc.CaKeyPath
	clientCertPath := mc.ClientCertPath
	clientKeyPath := mc.ClientKeyPath

	if caCertPath == "" {
		caCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca.pem")
	}

	if caKeyPath == "" {
		caKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "ca-key.pem")
	}

	if clientCertPath == "" {
		clientCertPath = filepath.Join(mcndirs.GetMachineCertDir(), "cert.pem")
	}

	if clientKeyPath == "" {
		clientKeyPath = filepath.Join(mcndirs.GetMachineCertDir(), "key.pem")
	}

	return cert.PathInfo{
		CaCertPath:       caCertPath,
		CaPrivateKeyPath: caKeyPath,
		ClientCertPath:   clientCertPath,
		ClientKeyPath:    clientKeyPath,
	}
}

func getSwarmOptions(configSwarm swarm.Options) swarm.Options {
	var swarmOpt swarm.Options

	if (reflect.DeepEqual(configSwarm, swarm.Options{})) {
		swarmOpt.IsSwarm = false
		return swarmOpt
	}
	swarmOpt.IsSwarm = true
	swarmOpt.Master = configSwarm.Master
	swarmOpt.Discovery = configSwarm.Discovery
	swarmOpt.Address = configSwarm.Address
	swarmOpt.ArbitraryFlags = configSwarm.ArbitraryFlags
	swarmOpt.Image = configSwarm.Image
	swarmOpt.Strategy = configSwarm.Strategy
	swarmOpt.Host = configSwarm.Host

	if configSwarm.Image == "" {
		swarmOpt.Image = "swarm:latest"
	}
	if configSwarm.Strategy == "" {
		swarmOpt.Strategy = "spread"
	}
	if configSwarm.Host == "" {
		swarmOpt.Host = "tcp://0.0.0.0:3376"
	}

	return swarmOpt
}

// yaml2cmd converts yaml tags with "_" to cmd tags "-", upto 5 occurance
func yaml2cmd(yamlTag string) string {
	return strings.Replace(yamlTag, "_", "-", 5)
}

func createMachine(driver drivers.Driver, node machineConfig, h *host.Host, api libmachine.API) error {
	mcnFlags := driver.GetCreateFlags()
	driverOpts := getDriverOptions(node, mcnFlags)

	h.HostOptions.AuthOptions.SkipCertGeneration = false

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := api.Create(h); err != nil {
		return fmt.Errorf("Error creating machine: %s", err)
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save store: %s", err)
	}

	return nil
}

func startMachine(driver drivers.Driver, node machineConfig, h *host.Host, api libmachine.API) error {
	mcnFlags := driver.GetCreateFlags()
	driverOpts := getDriverOptions(node, mcnFlags)

	h.HostOptions.AuthOptions.SkipCertGeneration = true

	if err := h.Driver.SetConfigFromFlags(driverOpts); err != nil {
		return fmt.Errorf("Error setting machine configuration from flags provided: %s", err)
	}

	if err := api.Start(h); err != nil {
		return fmt.Errorf("Error while starting machine: %s", err)
	}

	if err := api.Save(h); err != nil {
		return fmt.Errorf("Error attempting to save store: %s", err)
	}

	return nil
}
