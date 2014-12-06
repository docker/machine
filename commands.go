package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"text/tabwriter"
	"path"
	"path/filepath"
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"

	"github.com/docker/machine/drivers"
	_ "github.com/docker/machine/drivers/azure"
	_ "github.com/docker/machine/drivers/digitalocean"
	_ "github.com/docker/machine/drivers/none"
	_ "github.com/docker/machine/drivers/virtualbox"
)

type DockerCli struct{}

func (cli *DockerCli) getMethod(args ...string) (func(...string) error, bool) {
	camelArgs := make([]string, len(args))
	for i, s := range args {
		if len(s) == 0 {
			return nil, false
		}
		camelArgs[i] = strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
	}
	methodName := "Cmd" + strings.Join(camelArgs, "")
	method := reflect.ValueOf(cli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

func (cli *DockerCli) Cmd(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			return method(args[2:]...)
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
			return cli.CmdHelp()
		}
		return method(args[1:]...)
	}
	return cli.CmdHelp()
}

func (cli *DockerCli) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		options := ""
		if flags.FlagCountUndeprecated() > 0 {
			options = "[OPTIONS] "
		}
		fmt.Fprintf(os.Stderr, "\nUsage: docker %s %s%s\n\n%s\n\n", name, options, signature, description)
		flags.PrintDefaults()
		os.Exit(2)
	}
	return flags
}

func (cli *DockerCli) CmdHelp(args ...string) error {
	if len(args) > 1 {
		method, exists := cli.getMethod(args[:2]...)
		if exists {
			method("--help")
			return nil
		}
	}
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Fprintf(os.Stderr, "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}

	flag.Usage()

	return nil
}

func (cli *DockerCli) CmdLs(args ...string) error {
	cmd := cli.Subcmd("machines ls", "", "List machines")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display names")

	if err := cmd.Parse(args); err != nil {
		return err
	}

	store := NewStore()

	hostList, err := store.List()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 1, 3, ' ', 0)

	if !*quiet {
		fmt.Fprintln(w, "NAME\tACTIVE\tDRIVER\tSTATE\tURL")
	}

	wg := sync.WaitGroup{}

	for _, host := range hostList {
		host := host
		if *quiet {
			fmt.Fprintf(w, "%s\n", host.Name)
		} else {
			wg.Add(1)
			go func() {
				currentState, err := host.Driver.GetState()
				if err != nil {
					log.Errorf("error getting state for host %s: %s", host.Name, err)
				}

				url, err := host.GetURL()
				if err != nil {
					if err == drivers.ErrHostIsNotRunning {
						url = ""
					} else {
						log.Errorf("error getting URL for host %s: %s", host.Name, err)
					}
				}

				isActive, err := store.IsActive(&host)
				if err != nil {
					log.Errorf("error determining whether host %q is active: %s",
						host.Name, err)
				}

				activeString := ""
				if isActive {
					activeString = "*"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					host.Name, activeString, host.Driver.DriverName(), currentState, url)
				wg.Done()
			}()
		}
	}

	wg.Wait()
	w.Flush()

	return nil
}

func (cli *DockerCli) CmdCreate(args ...string) error {
	cmd := cli.Subcmd("machines create", "NAME", "Create machines")

	driverDesc := fmt.Sprintf(
		"Driver to create machine with. Available drivers: %s",
		strings.Join(drivers.GetDriverNames(), ", "),
	)

	driver := cmd.String([]string{"d", "-driver"}, "none", driverDesc)

	createFlags := drivers.RegisterCreateFlags(cmd)

	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() != 1 {
		cmd.Usage()
		return nil
	}

	name := cmd.Arg(0)

	store := NewStore()

	driverCreateFlags, _ := createFlags[*driver]
	host, err := store.Create(name, *driver, driverCreateFlags)
	if err != nil {
		return err
	}
	if err := store.SetActive(host); err != nil {
		return err
	}
	log.Infof("%q has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity", name)
	return nil
}

func (cli *DockerCli) CmdStart(args ...string) error {
	cmd := cli.Subcmd("machines start", "NAME", "Start a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Start()
}

func (cli *DockerCli) CmdStop(args ...string) error {
	cmd := cli.Subcmd("machines stop", "NAME", "Stop a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Stop()
}

func (cli *DockerCli) CmdRm(args ...string) error {
	cmd := cli.Subcmd("machines rm", "NAME", "Remove a machine")
	force := cmd.Bool([]string{"f", "-force"}, false, "Remove local configuration even if machine cannot be removed")

	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	isError := false

	store := NewStore()
	for _, host := range cmd.Args() {
		host := host
		if err := store.Remove(host, *force); err != nil {
			log.Errorf("Error removing machine %s: %s", host, err)
			isError = true
		}
	}
	if isError {
		return fmt.Errorf("There was an error removing a machine. To force remove it, pass the -f option. Warning: this might leave it running on the provider.")
	}
	return nil
}

func (cli *DockerCli) CmdIp(args ...string) error {
	cmd := cli.Subcmd("machines ip", "NAME", "Get the IP address of a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() > 1 {
		cmd.Usage()
		return nil
	}

	var (
		err   error
		host  *Host
		store = NewStore()
	)

	if cmd.NArg() == 1 {
		host, err = store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
	} else {
		host, err = store.GetActive()
		if err != nil {
			return err
		}
		if host == nil {
			os.Exit(1)
		}
	}

	ip, err := host.Driver.GetIP()
	if err != nil {
		return err
	}

	fmt.Println(ip)

	return nil
}

func (cli *DockerCli) CmdUrl(args ...string) error {
	cmd := cli.Subcmd("machines url", "NAME", "Get the URL of a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() > 1 {
		cmd.Usage()
		return nil
	}

	var (
		err   error
		host  *Host
		store = NewStore()
	)

	if cmd.NArg() == 1 {
		host, err = store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
	} else {
		host, err = store.GetActive()
		if err != nil {
			return err
		}
		if host == nil {
			os.Exit(1)
		}
	}

	url, err := host.GetURL()
	if err != nil {
		return err
	}

	fmt.Println(url)

	return nil
}

func (cli *DockerCli) CmdRestart(args ...string) error {
	cmd := cli.Subcmd("machines restart", "NAME", "Restart a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Driver.Restart()
}

func (cli *DockerCli) CmdKill(args ...string) error {
	cmd := cli.Subcmd("machines kill", "NAME", "Kill a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}
	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}
	return host.Driver.Kill()
}

func (cli *DockerCli) CmdSsh(args ...string) error {
	cmd := cli.Subcmd("machines ssh", "NAME [COMMAND ...]", "Log into or run a command on a machine with SSH")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() < 1 {
		cmd.Usage()
		return nil
	}

	i := 1
	for i < len(os.Args) && os.Args[i-1] != cmd.Arg(0) {
		i++
	}

	store := NewStore()

	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	sshCmd, err := host.Driver.GetSSHCommand(os.Args[i:]...)
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (cli *DockerCli) CmdActive(args ...string) error {
	cmd := cli.Subcmd("machines active", "[NAME]", "Get or set the active machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	store := NewStore()

	if cmd.NArg() == 0 {
		host, err := store.GetActive()
		if err != nil {
			return err
		}
		if host != nil {
			fmt.Println(host.Name)
		}
	} else if cmd.NArg() == 1 {
		host, err := store.Load(cmd.Arg(0))
		if err != nil {
			return err
		}
		if err := store.SetActive(host); err != nil {
			return err
		}
	} else {
		cmd.Usage()
	}

	return nil

}

func (cli * DockerCli) CmdExport(args ...string) error {
	cmd := cli.Subcmd("machines export", "[NAME]", "Export a machine to a tarfile")
		if err := cmd.Parse(args); err != nil {
		return err
	}

	store := NewStore()

	if  cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	hostName := cmd.Arg(0)
	hostPath := path.Join(store.Path, hostName)

	files, err := ioutil.ReadDir(hostPath)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	for _, file := range files {

		hdr := &tar.Header{
			Name: file.Name(),
			Mode: int64(file.Mode()),
			Size: file.Size(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		fileO, err := os.Open(path.Join(hostPath, file.Name()))
		if err != nil{
			return err
		}

		_, err = io.Copy(tw, fileO)
		if err != nil{
			return err
		}
	}

	output := hostName + ".tar"
	err = ioutil.WriteFile(output, buf.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil

}

func (cli * DockerCli) CmdImport(args ...string) error {
	cmd := cli.Subcmd("machines import", "[TARFILE]", "Import a machine from a tarfile")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if  cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	store := NewStore()

	tarName := cmd.Arg(0)
	hostName := strings.TrimSuffix(tarName, filepath.Ext(tarName))

	curDir, err := os.Getwd()
	tarPath := path.Join(curDir, tarName)
	data, err := ioutil.ReadFile(tarPath)
	if err != nil {
		return err
	}

	r := bytes.NewReader(data)
	tr := tar.NewReader(r)

	os.MkdirAll(path.Join(store.Path, hostName), 0700)

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, tr); err != nil {
			return err
		}

		filePath := path.Join(store.Path, hostName, hdr.Name)
		err = ioutil.WriteFile(filePath, buf.Bytes(), 0600)
		if err != nil {
			return err
		}
	}

	return nil
}


func (cli *DockerCli) CmdInspect(args ...string) error {
	cmd := cli.Subcmd("machines inspect", "[NAME]", "Get detailed information about a machine")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	store := NewStore()
	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	prettyJson, err := json.MarshalIndent(host, "", "    ")
	if err != nil {
		return err
	}

	fmt.Println(string(prettyJson))

	return nil
}

func (cli *DockerCli) CmdUpgrade(args ...string) error {
	cmd := cli.Subcmd("machines upgrade", "[NAME]", "Upgrade a machine to the latest version of Docker")
	if err := cmd.Parse(args); err != nil {
		return err
	}

	if cmd.NArg() == 0 {
		cmd.Usage()
		return nil
	}

	store := NewStore()
	host, err := store.Load(cmd.Arg(0))
	if err != nil {
		return err
	}

	return host.Driver.Upgrade()
}
