package localbinary

import (
	"bufio"
	"fmt"
	"io"
	"testing"
	"time"

	"os"

	"github.com/docker/machine/libmachine/log"
)

type FakeExecutor struct {
	stdout, stderr io.ReadCloser
	closed         bool
}

func (fe *FakeExecutor) Start() (*bufio.Scanner, *bufio.Scanner, error) {
	return bufio.NewScanner(fe.stdout), bufio.NewScanner(fe.stderr), nil
}

func (fe *FakeExecutor) Close() error {
	fe.closed = true
	return nil
}

func TestLocalBinaryPluginAddress(t *testing.T) {
	lbp := &Plugin{}
	expectedAddr := "127.0.0.1:12345"

	lbp.addrCh = make(chan string, 1)
	lbp.addrCh <- expectedAddr

	// Call the first time to read from the channel
	addr, err := lbp.Address()
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	if addr != expectedAddr {
		t.Fatal("Expected did not match actual address")
	}

	// Call the second time to read the "cached" address value
	addr, err = lbp.Address()
	if err != nil {
		t.Fatalf("Expected no error, instead got %s", err)
	}
	if addr != expectedAddr {
		t.Fatal("Expected did not match actual address")
	}
}

func TestLocalBinaryPluginAddressTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test")
	}
	lbp := &Plugin{}
	lbp.addrCh = make(chan string, 1)
	go func() {
		_, err := lbp.Address()
		if err == nil {
			t.Fatalf("Expected to get a timeout error, instead got %s", err)
		}
	}()
	time.Sleep(defaultTimeout + 1)
}

func TestLocalBinaryPluginClose(t *testing.T) {
	lbp := &Plugin{}
	lbp.stopCh = make(chan bool, 1)
	go lbp.Close()
	stopped := <-lbp.stopCh
	if !stopped {
		t.Fatal("Close did not send a stop message on the proper channel")
	}
}

func TestExecServer(t *testing.T) {
	log.SetDebug(true)
	machineName := "test"

	logReader, logWriter := io.Pipe()

	log.SetOutput(logWriter)

	defer func() {
		log.SetDebug(false)
		log.SetOutput(os.Stderr)
	}()

	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	fe := &FakeExecutor{
		stdout: stdoutReader,
		stderr: stderrReader,
	}

	lbp := &Plugin{
		MachineName: machineName,
		Executor:    fe,
		addrCh:      make(chan string, 1),
		stopCh:      make(chan bool, 1),
	}

	finalErr := make(chan error)

	// Start the docker-machine-foo plugin server
	go func() {
		finalErr <- lbp.execServer()
	}()

	expectedAddr := "127.0.0.1:12345"
	expectedPluginOut := "Doing some fun plugin stuff..."
	expectedPluginErr := "Uh oh, something in plugin went wrong..."

	logScanner := bufio.NewScanner(logReader)

	if _, err := io.WriteString(stdoutWriter, expectedAddr+"\n"); err != nil {
		t.Fatalf("Error attempting to write plugin address: %s", err)
	}

	if addr := <-lbp.addrCh; addr != expectedAddr {
		t.Fatalf("Expected to read the expected address properly in server but did not")
	}

	expectedOut := fmt.Sprintf("%s%s", fmt.Sprintf(pluginOutPrefix, machineName), expectedPluginOut)

	if _, err := io.WriteString(stdoutWriter, expectedPluginOut+"\n"); err != nil {
		t.Fatalf("Error attempting to write to out in plugin: %s", err)
	}

	if logScanner.Scan(); logScanner.Text() != expectedOut {
		t.Fatalf("Output written to log was not what we expected\nexpected: %s\nactual:   %s", expectedOut, logScanner.Text())
	}

	expectedErr := fmt.Sprintf("%s%s", fmt.Sprintf(pluginErrPrefix, machineName), expectedPluginErr)

	if _, err := io.WriteString(stderrWriter, expectedPluginErr+"\n"); err != nil {
		t.Fatalf("Error attempting to write to err in plugin: %s", err)
	}

	if logScanner.Scan(); logScanner.Text() != expectedErr {
		t.Fatalf("Error written to log was not what we expected\nexpected: %s\nactual:   %s", expectedErr, logScanner.Text())
	}

	lbp.Close()

	if err := <-finalErr; err != nil {
		t.Fatalf("Error serving: %s", err)
	}
}
