package plugin

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/drivers/rpc"
)

var (
	heartbeatTimeout = 500 * time.Millisecond
)

func RegisterDriver(d drivers.Driver) {
	if os.Getenv(localbinary.PluginEnvKey) != localbinary.PluginEnvVal {
		fmt.Fprintln(os.Stderr, `This is a Docker Machine plugin binary.
Plugin binaries are not intended to be invoked directly.
Please use this plugin through the main 'docker-machine' binary.`)
		os.Exit(1)
	}

	libmachine.SetDebug(true)

	rpcd := rpcdriver.NewRpcServerDriver(d)
	rpc.Register(rpcd)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading RPC server: %s\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println(listener.Addr())

	go http.Serve(listener, nil)

	for {
		select {
		case <-rpcd.CloseCh:
			os.Exit(0)
		case <-rpcd.HeartbeatCh:
			continue
		case <-time.After(heartbeatTimeout):
			os.Exit(1)
		}
	}
}
