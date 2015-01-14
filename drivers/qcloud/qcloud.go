package qcloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	//"github.com/docker/docker/api"
	"github.com/docker/machine/drivers"
	qapi "github.com/docker/machine/drivers/qcloud/QcloudApi"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

// Driver is the driver used when no driver is selected. It is used to
// connect to existing Docker hosts by specifying the URL of the host as
// an option.
type Driver struct {
	//URL string
	Region      string
	Secretid    string
	Secretkey   string
	ImageType   int
	ImageId     int
	Cpu         int
	Mem         int
	Bandwith    int
	StorageType int
	StorageSize int
	Period      int
	storePath   string
	SSHKeyPath  string
	//set below params in Create()
	InstanceId string
	SSHKeyID   int
	IPAddress  string
	UserName   string //ubuntu is ubunt, centOS root
	Debug      bool
}

func init() {
	drivers.Register("qcloud", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "qcloud-region",
			Usage: "qcloud region, default gz",
			Value: "gz",
		},
		cli.StringFlag{
			Name:  "qcloud-secretid",
			Usage: "qcloud secretid",
			Value: "",
		},
		cli.StringFlag{
			Name:  "qcloud-secretkey",
			Usage: "qcloud secretkey",
			Value: "",
		},
		cli.IntFlag{
			Name:  "qcloud-instance-type",
			Usage: "qcloud instance type, default 1",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-image-type",
			Usage: "1 private images, 2 public images, default 2",
			Value: 2,
		},
		cli.IntFlag{
			Name:  "qcloud-image-id",
			Usage: "qcloud image id, ref http://wiki.qcloud.com/wiki/, default ubuntu12.04 docker 64bit",
			Value: 9, //ubuntu12.04 docker 64bit
		},
		cli.IntFlag{
			Name:  "qcloud-cpu",
			Usage: "qcloud cpu num, default 1",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-mem",
			Usage: "qcloud mem size in GB, default 1",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-bandwidth",
			Usage: "qcloud bandwidth in Mbps, 0 for none, default 1",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-storage-type",
			Usage: "1 local disk, 2 cloud disk, default 1",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-storage-size",
			Usage: "disk size in GB, 0 for none, default 0",
			Value: 0,
		},
		cli.IntFlag{
			Name:  "qcloud-period",
			Usage: "charge period in month",
			Value: 1,
		},
		cli.IntFlag{
			Name:  "qcloud-sshkey-id",
			Usage: "sshkey id in qcloud, only rsa supported so far",
			Value: 0,
		},
		cli.StringFlag{
			Name:  "qcloud-sshkey-path",
			Usage: "ssh privatekey path, download this key from qcloud and save it here, default ~/.ssh/id_rsa",
			Value: filepath.Join(drivers.GetHomeDir(), ".ssh", "id_rsa"),
		},
		cli.IntFlag{
			Name:  "qcloud-debug",
			Usage: "Debug mode will print more info",
			Value: 0,
		},
	}
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "qcloud"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Region = flags.String("qcloud-region")
	d.Secretid = flags.String("qcloud-secretid")
	d.Secretkey = flags.String("qcloud-secretkey")
	d.ImageType = flags.Int("qcloud-image-type")
	d.ImageId = flags.Int("qcloud-image-id")
	d.Cpu = flags.Int("qcloud-cpu")
	d.Mem = flags.Int("qcloud-mem")
	d.Bandwith = flags.Int("qcloud-bandwidth")
	d.StorageType = flags.Int("qcloud-storage-type")
	d.StorageSize = flags.Int("qcloud-storage-size")
	d.Period = flags.Int("qcloud-period")
	d.SSHKeyID = flags.Int("qcloud-sshkey-id")
	d.SSHKeyPath = flags.String("qcloud-sshkey-path")
	d.UserName = "ubuntu"
	d.Debug = flags.Bool("qcloud-debug")
	///*
	if d.Secretid == "" {
		return fmt.Errorf("qcloud driver requires the --qcloud-secretid option")
	}

	if d.Secretkey == "" {
		return fmt.Errorf("qcloud driver requires the --qcloud-secretkey option")
	}

	if d.SSHKeyID == 0 {
		return fmt.Errorf("qcloud driver requires the --qcloud-sshkey-id option")
	}
	//*/

	return nil
}

func (d *Driver) GetURL() (string, error) {
	//return d.URL, nil
	//return "", nil
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	config := map[string]interface{}{
		"secretId":  d.Secretid,
		"secretKey": d.Secretkey}
	//"debug":     d.Debug}
	params := map[string]interface{}{
		"Region":        d.Region,
		"Action":        "DescribeInstances",
		"instanceIds.1": d.InstanceId,
		"limit":         1}

	retData, err :=
		qapi.SendRequest("cvm", params, config)

	if err != nil {
		return state.None, fmt.Errorf("query qcloud vm:%s status error:%s", d.InstanceId, err)
	}

	//fmt.Printf("http_res:%v\n", retData)
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		return state.None, fmt.Errorf("qcloud res not valid json:%s", retData)
	}

	instanceSet, is_ok :=
		jsonObj["instanceSet"].([]interface{})
	if !is_ok || len(instanceSet) == 0 {
		return state.None, fmt.Errorf("qcloud not contain any instanc:%s", jsonObj)
	}

	instance, is_ok :=
		instanceSet[0].(map[string]interface{})
	if !is_ok {
		return state.None, fmt.Errorf("qcloud res contain no valid instanc:%s", instanceSet)
	}

	var qstatus state.State
	qstatus_float, is_ok := instance["status"].(float64)
	if !is_ok {
		return state.None, fmt.Errorf("qcloud res instance contain no valid status:%s", instance)
	}
	qstatus = state.State(qstatus_float)

	switch qstatus {
	case 2:
		return state.Running, nil
	case 4:
		return state.Stopped, nil
	case 9:
		return state.Stopping, nil
	case 8:
		return state.Starting, nil
	case 1:
		return state.Error, nil
	}

	//fmt.Printf("qstatus:%v\n", qstatus)
	return state.None, nil
}

func (d *Driver) Create() error {

	// create instance
	log.Infof("Waiting for launching instance...")
	config := map[string]interface{}{
		"secretId":  d.Secretid,
		"secretKey": d.Secretkey,
		"debug":     d.Debug}
	params := map[string]interface{}{
		"Action":       "RunInstances",
		"Region":       d.Region,
		"instanceType": 1,
		"imageType":    d.ImageType,
		"imageId":      d.ImageId,
		"cpu":          d.Cpu,
		"mem":          d.Mem,
		"goodsNum":     1,
		"period":       d.Period,
		"bandwidth":    d.Bandwith,
		"storageType":  d.StorageType,
		"storageSize":  d.StorageSize}
	var err error
	var retData string
	var jsonObj map[string]interface{}

	retData, err = qapi.SendRequest("cvm", params, config)
	if d.Debug {
		log.Debugf("retData[%v]", retData)
	}
	if err != nil {
		fmt.Errorf("create qcloud error:%s", err)
		return err
	}

	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		fmt.Errorf("qcloud res not valid json:%s", retData)
		return err
	}
	retCode := 0
	retCode = int(jsonObj["code"].(float64))
	if retCode != 0 {
		err = errors.New(fmt.Sprintf("%v", jsonObj["message"]))
		return err
	}
	log.Infof("Create instance request send OK.")
	log.Infof("Waiting for produce instance ...")

queryDeal:
	// Query deal for instanceId
	log.Infof("Query instance ...")
	deals := jsonObj["dealIds"].([]interface{})
	params = map[string]interface{}{
		"Action":    "DescribeDeals",
		"dealIds.1": deals[0]}
	retData, err = qapi.SendRequest("trade", params, config)
	if err != nil {
		fmt.Errorf("query qcloud deal error.", err)
		return err
	}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		fmt.Errorf("qcloud res not valid json:%s", retData)
		return err
	}
	retCode = int(jsonObj["code"].(float64))
	if retCode != 0 {
		err = errors.New(fmt.Sprintf("%v", jsonObj["message"]))
		return err
	}
	dealDetails, _ := jsonObj["dealDetails"].([]interface{})
	dealDetail, _ := dealDetails[0].(map[string]interface{})
	goodsDetail, _ := dealDetail["goodsDetail"].(map[string]interface{})
	instanceIds := goodsDetail["instanceId"].([]interface{})
	if len(instanceIds) == 0 {
		log.Infof("Instance not ready yet...")
		time.Sleep(30 * time.Second)
		goto queryDeal
	}
	instanceId := instanceIds[0]
	d.InstanceId = fmt.Sprintf("%v", instanceId)

	log.Infof("Instance instanceId[%v]", instanceId)
	d.waitForInstance()
	log.Infof("Instance is ready")

bindSSHKey:
	// bind sshkey
	log.Infof("Waiting for bind SSH key to instance...")
	params = map[string]interface{}{
		"Action":      "BindCvmSecretKey",
		"instanceId":  instanceId,
		"secretKeyId": d.SSHKeyID}
	retData, err = qapi.SendRequest("cvm", params, config)
	if err != nil {
		fmt.Errorf("query qcloud deal error.", err)
		return err
	}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		fmt.Errorf("qcloud res not valid json:%s", retData)
		return err
	}
	retCode = int(jsonObj["code"].(float64))
	if retCode != 0 {
		err = errors.New(fmt.Sprintf("%v", jsonObj["message"]))
		return err
	}
	requestId := int(jsonObj["requestId"].(float64))
	time.Sleep(20 * time.Second)

bindSSHKeyRequestQuery:
	// query bind result
	params = map[string]interface{}{
		"Action":    "DescribeTaskResponse",
		"requestId": requestId}
	retData, err = qapi.SendRequest("cvm", params, config)
	if err != nil {
		fmt.Errorf("query qcloud deal error.", err)
		return err
	}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		fmt.Errorf("qcloud res not valid json:%s", retData)
		return err
	}
	retCode = int(jsonObj["code"].(float64))
	if retCode != 0 {
		err = errors.New(fmt.Sprintf("%v", jsonObj["message"]))
		return err
	}
	detail := jsonObj["detail"].(map[string]interface{})
	bindSSHKeyRet := int(detail["resCode"].(float64))
	if bindSSHKeyRet != 0 {
		time.Sleep(30 * time.Second)
		if bindSSHKeyRet == 2 {
			log.Infof("Bind SSH key in process ret[%d], query again ......", bindSSHKeyRet)
			goto bindSSHKeyRequestQuery
		}
		log.Infof("Bind SSH key error ret[%d], try again ......", bindSSHKeyRet)
		goto bindSSHKey
	}

	log.Infof("Bind SSH key to instance OK.")

	log.Infof("Waiting for query instance info ...")
	d.waitForInstance()
	// query instance
	params = map[string]interface{}{
		"Region":        d.Region,
		"Action":        "DescribeInstances",
		"instanceIds.1": instanceId,
		"limit":         1}
	retData, err = qapi.SendRequest("cvm", params, config)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		return err
	}
	instanceSet, is_ok := jsonObj["instanceSet"].([]interface{})
	if !is_ok || len(instanceSet) == 0 {
		return err
	}
	instance, is_ok := instanceSet[0].(map[string]interface{})
	if !is_ok {
		fmt.Errorf("qcloud res contain no valid instanc:%s", instanceSet)
		return err
	}
	wanIpSet, is_ok := instance["wanIpSet"].([]interface{})
	if !is_ok || len(wanIpSet) == 0 {
		return err
	}
	wanIp := wanIpSet[0]
	d.IPAddress = fmt.Sprintf("%v", wanIp)
	log.Infof("Query instance info OK. wanIp[%v].", wanIp)

	cmdString := "whoami"
	log.Infof("Try to run a SSH cmd[%v].", cmdString)
	cmd, err := d.GetSSHCommand(cmdString)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	/*
		// upload machine pubkey
		log.Debugf("Updating /etc/default/docker to use identity auth...")

		cmd, err := d.GetSSHCommand("echo 'export DOCKER_OPTS=\"--auth=identity --host=unix:// --host=tcp://0.0.0.0:2376 --auth-authorized-dir=/root/.docker/authorized-keys.d\"' | sudo tee -a /etc/default/docker")
		if err != nil {
			return err
		}
		if err := cmd.Run(); err != nil {
			return err
		}

		log.Debugf("Adding key to authorized-keys.d...")
		if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/root/.docker/authorized-keys.d"); err != nil {
			return err
		}
	*/

	log.Infof("Done.")

	return nil
}

func (d *Driver) Start() error {
	config := map[string]interface{}{
		"secretId":  d.Secretid,
		"secretKey": d.Secretkey,
		"debug":     false}
	params := map[string]interface{}{
		"Action":        "StartInstances",
		"instanceIds.1": d.InstanceId}

	retData, err :=
		qapi.SendRequest("cvm", params, config)

	if err != nil {
		return fmt.Errorf("start qcloud vm:%s error:%s", d.InstanceId, err)
	}

	//fmt.Print("http_res:", retData)
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		return fmt.Errorf("qcloud res not valid json:%s", retData)
	}

	//fmt.Printf("jsonObj:%v", jsonObj)
	all_ok := isResAllOk(jsonObj)
	if all_ok {
		return nil
	} else {
		return fmt.Errorf("qcloud Start fail:%v", jsonObj)
	}
}

func (d *Driver) Stop() error {
	config := map[string]interface{}{
		"secretId":  d.Secretid,
		"secretKey": d.Secretkey,
		"debug":     false}
	params := map[string]interface{}{
		"Action":        "StopInstances",
		"instanceIds.1": d.InstanceId}

	retData, err :=
		qapi.SendRequest("cvm", params, config)

	if err != nil {
		return fmt.Errorf("stop qcloud vm:%s error:%s", d.InstanceId, err)
	}

	//fmt.Print("http_res:", retData)
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		return fmt.Errorf("qcloud res not valid json:%s", retData)
	}

	//fmt.Printf("jsonObj:%v", jsonObj)
	all_ok := isResAllOk(jsonObj)
	if all_ok {
		return nil
	} else {
		return fmt.Errorf("qcloud Stop fail:%v", jsonObj)
	}
}

func (d *Driver) Remove() error {
	return d.Stop()
}

func (d *Driver) Restart() error {
	config := map[string]interface{}{
		"secretId":  d.Secretid,
		"secretKey": d.Secretkey,
		"debug":     false}
	params := map[string]interface{}{
		"Action":        "RestartInstances",
		"instanceIds.1": d.InstanceId}

	retData, err :=
		qapi.SendRequest("cvm", params, config)

	if err != nil {
		return fmt.Errorf("restart qcloud vm:%s error:%s", d.InstanceId, err)
	}

	//fmt.Print("http_res:", retData)
	var jsonObj map[string]interface{}
	err = json.Unmarshal([]byte(retData), &jsonObj)
	if err != nil {
		return fmt.Errorf("qcloud res not valid json:%s", retData)
	}

	//fmt.Printf("jsonObj:%v", jsonObj)
	all_ok := isResAllOk(jsonObj)
	if all_ok {
		return nil
	} else {
		return fmt.Errorf("qcloud Restart fail:%v", jsonObj)
	}
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) Upgrade() error {
	sshCmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install -y lxc-docker")
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	return sshCmd.Run()
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, d.UserName, d.publicSSHKeyPath(), args...), nil
}

func (d *Driver) publicSSHKeyPath() string {
	// @TODO check file and login
	return d.SSHKeyPath
}

func isResAllOk(jsonObj map[string]interface{}) bool {

	detailSet, is_ok :=
		jsonObj["detail"].(map[string]interface{})
	if !is_ok || len(detailSet) == 0 {
		//fmt.Printf("qcloud res not contain any instanc:%s\n", jsonObj)
		return false
	}

	all_ok := true
	for _, v := range detailSet {
		detail, is_ok := v.(map[string]interface{})
		if !is_ok {
			all_ok = false
			//fmt.Printf("qcloud res format error:%v\n", v)
			continue
		}

		errorCode, is_ok := detail["code"].(float64)
		if !is_ok {
			//fmt.Printf("qcloud res contain no code:%v\n", detail)
			all_ok = false
			continue
		}

		if errorCode != 0 {
			//fmt.Printf("qcloud operation failed:%v\n", detail)
			all_ok = false
			continue
		}
	}

	return all_ok
}

func (d *Driver) waitForInstance() error {
	for {
		st, err := d.GetState()
		if err != nil {
			return err
		}
		if st == state.Running {
			break
		}
		time.Sleep(10 * time.Second)
	}

	return nil
}
