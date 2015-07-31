package triton

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	driverName                  = "triton"
	TritonDefaultDockerPort     = 2376
	TritonDefaultCloudapiDomain = "api.joyent.com"
)

type Driver struct {
	CloudApiURL   string
	DockerApiURL  string
	DataCenter    string
	Account       string
	PrivateKey    string
	SkipTlsVerify bool
	MachineName   string
	storePath     string
}

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "triton-url",
			Usage:  "Triton cloudapi URL",
			Value:  "",
			EnvVar: "SDC_URL",
		},
		cli.StringFlag{
			Name:   "triton-datacenter",
			Usage:  "Triton datacenter name",
			Value:  "us-east-1",
			EnvVar: "SDC_DC",
		},
		cli.StringFlag{
			Name:   "triton-account",
			Usage:  "Triton account name",
			Value:  "",
			EnvVar: "SDC_ACCOUNT",
		},
		cli.StringFlag{
			Name:   "triton-key",
			Usage:  "SSH private key for Triton authentication",
			Value:  "",
			EnvVar: "SDC_KEY",
		},
		cli.BoolFlag{
			Name:   "triton-skip-tls-verify",
			Usage:  "Skip tls verification 'true' or 'false' (defaults to 'false')",
			EnvVar: "SDC_SKIP_TLS_VERIFY",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{
		MachineName: machineName,
		storePath:   storePath,
	}, nil
}

/* --------------------------------------------------------- */
/* Implement the drivers.Driver interface.                   */
/* --------------------------------------------------------- */

// AuthorizePort authorizes a port for machine access
func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	log.Infof("Generating %s user certificates - you will be prompted for", driverName)
	log.Infof("your SSH private key password (if it's password protected).")

	err := d.RegisterWithSdcCloudApi()
	if err != nil {
		return err
	}

	err = d.GenerateCertificates()
	if err != nil {
		return err
	}

	log.Infof("Success!")

	return nil
}

// DeauthorizePort removes a port for machine access
func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

// DriverName returns the name of the driver as it is registered
func (d *Driver) DriverName() string {

	/**
	 * Overriding the driver name to avoid SSH provisioning - issue #886
	 *
	 * We only need override the driver name in the 'create' step, so this
	 * approximately checks if this is a create.
	 */
	for _, v := range os.Args {
		if v == "create" {
			return "none"
		}
	}

	return driverName
}

// GetIP returns an IP or hostname that this host is available at
// e.g. 1.2.3.4 or docker-host-d60b70a14d3a.cloudapp.net
func (d *Driver) GetIP() (string, error) {
	// DockerApiURL looks like: 'tcp://foo.bar:2376'
	u, err := url.Parse(d.DockerApiURL)
	if err != nil {
		return "", err
	}
	return strings.Split(u.Host, ":")[0], nil
}

// GetMachineName returns the name of the machine
func (d *Driver) GetMachineName() string {
	return d.MachineName
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return "", fmt.Errorf("SSH is not available for the triton driver")
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	return 22, nil
}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	return d.Account
}

// GetSSHKeyPath returns key path for use with ssh
func (d *Driver) GetSSHKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

// GetURL returns a Docker compatible host URL for connecting to this host
// e.g. tcp://1.2.3.4:2376
func (d *Driver) GetURL() (string, error) {
	return d.DockerApiURL, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	u := d.GetHttpsURL()

	resp, err := client.Get(fmt.Sprintf("%s/_ping", u))
	if err != nil {
		return state.Error, nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if string(body[:]) == "OK" {
		return state.Running, nil
	}
	return state.None, nil
}

// Kill stops a host forcefully
func (d *Driver) Kill() error {
	return fmt.Errorf("Kill is not available for the triton driver")
}

// PreCreateCheck allows for pre-create operations to make sure a driver is ready for creation
func (d *Driver) PreCreateCheck() error {
	return nil
}

// Remove a host
func (d *Driver) Remove() error {
	return nil
}

// Restart a host. This may just call Stop(); Start() if the provider does not
// have any special restart behaviour.
func (d *Driver) Restart() error {
	return fmt.Errorf("Restart is not available for the triton driver")
}

// SetConfigFromFlags configures the driver with the object that was returned
// by RegisterCreateFlags
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.CloudApiURL = flags.String("triton-url")
	d.Account = flags.String("triton-account")
	d.DataCenter = flags.String("triton-datacenter")
	d.PrivateKey = flags.String("triton-key")
	d.SkipTlsVerify = flags.Bool("triton-skip-tls-verify")

	if d.CloudApiURL == "" {
		if d.DataCenter == "" {
			log.Fatal("You must specify a cloudapi url or datacenter name")
		}
		// Shortend format for the cloudapi name, e.g. "us-east-1"
		d.CloudApiURL = fmt.Sprintf("https://%s.%s", d.DataCenter, TritonDefaultCloudapiDomain)
	}

	if d.Account == "" {
		return fmt.Errorf("You must specify the account name using --triton-account")
	}

	if d.PrivateKey == "" {
		homedir := utils.GetHomeDir()
		if homedir == "" {
			return fmt.Errorf("You must specify the SSH key using --triton-key")
		}
		d.PrivateKey = path.Join(homedir, ".ssh", "id_rsa")
	}

	_, err := os.Stat(d.PrivateKey)
	if err != nil {
		return fmt.Errorf("Unable to access SSH key file %s", d.PrivateKey)
	}

	log.Debugf("CloudApiURL: %s", d.CloudApiURL)
	log.Debugf("Account: %s", d.Account)
	log.Debugf("DataCenter: %s", d.DataCenter)
	log.Debugf("PrivateKey: %s", d.PrivateKey)
	log.Debugf("SkipTlsVerify: %d", d.SkipTlsVerify)

	return nil
}

// Start a host
func (d *Driver) Start() error {
	return fmt.Errorf("Start is not available for the triton driver")
}

// Stop a host gracefully
func (d *Driver) Stop() error {
	return fmt.Errorf("Stop is not available for the triton driver")
}

/* --------------------------------------------------------- */
/* End of driver.Driver                                      */
/* --------------------------------------------------------- */

// GetHttpsURL returns a https url to the docker endpoint
// e.g. https://1.2.3.4:2376
func (d *Driver) GetHttpsURL() string {
	return strings.Replace(d.DockerApiURL, "tcp://", "https://", 1)
}

/*
 * Download the certificate authority file from the sdc-docker server.
 */
func (d *Driver) DownloadCa() error {
	dockerHttpsUrl := d.GetHttpsURL()
	caUrl := fmt.Sprintf("%s/ca.pem", dockerHttpsUrl)
	log.Debugf("Downloading ca.pem file from %s", caUrl)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: d.SkipTlsVerify},
		},
	}
	resp, err := client.Get(caUrl)
	if err != nil {
		log.Debugf("Unable to open http request to url: %s", caUrl)
		return err
	}
	defer resp.Body.Close()

	caFile := path.Join(utils.GetMachineDir(), d.MachineName, "ca.pem")
	out, err := os.Create(caFile)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return err
}

func RunCommand(cmd []string, stdin string) (string, string, error) {
	var stdinPipe io.WriteCloser
	var err error

	log.Debugf("runCommand %s stdin length %d", cmd, len(stdin))

	args := cmd[1:]
	subProcess := exec.Command(cmd[0], args...)

	if stdin != "" {
		stdinPipe, err = subProcess.StdinPipe()
		if err != nil {
			log.Fatal("runCommand subProcess.StdinPipe failed")
		}
	}

	var bufout bytes.Buffer
	var buferr bytes.Buffer
	subProcess.Stdout = &bufout
	subProcess.Stderr = &buferr

	err = subProcess.Start()
	if err != nil {
		log.Debugf("runCommand subProcess.Start failed")
		return "", "", err
	}

	if stdin != "" {
		io.WriteString(stdinPipe, stdin)
		stdinPipe.Close()
	}

	err = subProcess.Wait()
	stdout := bufout.String()
	stderr := buferr.String()

	if err != nil {
		log.Debugf("runCommand subProcess.Wait failed")
	}

	if stderr != "" {
		log.Debugf("Stderr: %s", stderr)
	}

	return stdout, stderr, err
}

func (d *Driver) MakeCloudApiRequest(now string, encDateString string, sshKeyId string) error {
	//response=$(curl $CURL_OPTS $curlOpts -isS \
	//    -H "Accept:application/json" -H "api-version:*" -H "Date: ${now}" \
	//    -H "Authorization: Signature keyId=\"/$account/keys/$sshKeyId\",algorithm=\"rsa-sha256\" ${signature}" \
	//    --url $cloudapiUrl/$account/services)
	cloudapiUrl := fmt.Sprintf("%s/%s/services", d.CloudApiURL, d.Account)
	log.Debugf("CloudAPI url: ", cloudapiUrl)

	req, err := http.NewRequest("GET", cloudapiUrl, nil)
	if err != nil {
		log.Debugf("http.NewRequest failed for url %s", cloudapiUrl)
		return err
	}

	sig := fmt.Sprintf("Signature keyId=\"/%s/keys/%s\",algorithm=\"rsa-sha256\" %s",
		d.Account, sshKeyId, encDateString)
	req.Header.Add("Authorization", sig)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("api-version", "*")
	req.Header.Add("Date", now)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: d.SkipTlsVerify},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("client.Do failed for url %s", cloudapiUrl)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("resp.Body.read failed for url %s", cloudapiUrl)
		return err
	}

	log.Debugf("CloudAPI response: %s", body)

	var respMap map[string]interface{}
	json.Unmarshal([]byte(body), &respMap)

	if resp.StatusCode == http.StatusForbidden { // 403
		log.Fatalf("ERROR: CloudAPI registration was forbidden: %s", respMap["message"])
	}

	if resp.StatusCode != http.StatusOK { // 200
		log.Fatalf("ERROR: CloudAPI registration failed: %s", respMap["message"])
	}

	dockerUrl, ok := respMap["docker"].(string)
	if !ok {
		log.Fatalf("Could not convert docker response url to string, respMap %s", respMap)
	}

	// Sanity check the url.
	_, err = url.Parse(dockerUrl)
	if err != nil {
		log.Fatalf("Cloudapi returned an invalid url: %s - %s", dockerUrl, err)
	}

	d.DockerApiURL = dockerUrl

	return nil
}

func (d *Driver) RegisterWithSdcCloudApi() error {
	log.Debugf("registering with sdc cloud api")

	// now=$(date -u "+%a, %d %h %Y %H:%M:%S GMT")
	now := time.Now().UTC().Format(time.RFC1123)

	// signature=$(echo -n ${now} | openssl dgst -sha256 -sign $sshPrivKeyPath | openssl enc -e -a | tr -d '\n')
	cmd := []string{"openssl", "dgst", "-sha256", "-sign", d.PrivateKey}
	stdout, _, err := RunCommand(cmd, now)
	if err != nil {
		log.Debugf("command %s failed %s", cmd, err)
		return err
	}

	cmd = []string{"openssl", "enc", "-e", "-a"}
	encDateString, _, err := RunCommand(cmd, stdout)
	if err != nil {
		log.Debugf("command %s failed %s", cmd, err)
		return err
	}

	//ssh-keygen -l -f "$sshPubKeyPath" | awk '{print $2}' | tr -d '\n';
	cmd = []string{"ssh-keygen", "-l", "-f", d.PrivateKey + ".pub"}
	stdout, _, err = RunCommand(cmd, "")
	if err != nil {
		log.Debugf("command %s failed %s", cmd, err)
		return err
	}
	sshKeyIdSplit := strings.SplitN(stdout, " ", 3)
	if len(sshKeyIdSplit) < 2 {
		log.Debugf("invalid ssh key id, length is %d", len(sshKeyIdSplit))
		return err
	}
	sshKeyId := sshKeyIdSplit[1]

	// Register this user/key with the SDC cloud API.
	err = d.MakeCloudApiRequest(now, encDateString, sshKeyId)
	if err != nil {
		log.Debugf("MakeCloudApiRequest failed %s", err)
		return err
	}

	return nil
}

/*
 * Generate the client certificates (from the users private SSH key).
 *
 * This also generates the server*.pem certificate files, but these are
 * not used by sdc-docker (it just keeps docker-machine happy).
 */
func (d *Driver) GenerateCertificates() error {
	err := d.DownloadCa()
	if err != nil {
		return err
	}

	var keyFile = path.Join(utils.GetMachineDir(), d.MachineName, "key.pem")
	var csrFile = path.Join(utils.GetMachineDir(), d.MachineName, "cert.csr")
	var certFile = path.Join(utils.GetMachineDir(), d.MachineName, "cert.pem")

	cmd := exec.Command("openssl", "rsa", "-in", d.PrivateKey, "-outform", "pem", "-out", keyFile)
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("openssl", "req", "-new", "-key", keyFile, "-out", csrFile, "-subj", "/CN="+d.Account)
	err = cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("openssl", "x509", "-req", "-days", "365", "-in", csrFile, "-signkey", keyFile, "-out", certFile)
	err = cmd.Run()
	if err != nil {
		return err
	}

	log.Debugf("Generating server certificates")

	pemBytes, _ := ioutil.ReadFile(keyFile)
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return errors.New("ssh: no key found")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(123),
		Subject: pkix.Name{
			CommonName: d.Account,
		},
		SignatureAlgorithm:    x509.SHA512WithRSA,
		PublicKeyAlgorithm:    x509.RSA,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        false,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)
	if err != nil {
		return err
	}

	log.Debugf("writing server certificates")

	var serverKeyFile = path.Join(utils.GetMachineDir(), d.MachineName, "server-key.pem")
	var serverCertFile = path.Join(utils.GetMachineDir(), d.MachineName, "server.pem")

	certOut, _ := os.OpenFile(serverCertFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer certOut.Close()

	keyOut, _ := os.OpenFile(serverKeyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b})
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})

	return nil
}
