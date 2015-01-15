package amz

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/docker/machine/state"
	awsauth "github.com/smartystreets/go-aws-auth"
)

type (
	EC2 struct {
		Endpoint string
		Auth     Auth
		Region   string
	}

	Instance struct {
		info EC2Instance
	}

	EC2Instance struct {
		InstanceId    string `xml:"instanceId"`
		ImageId       string `xml:"imageId"`
		InstanceState struct {
			Code int    `xml:"code"`
			Name string `xml:"name"`
		} `xml:"instanceState"`
		PrivateDnsName string `xml:"privateDnsName"`
		DnsName        string `xml:"dnsName"`
		Reason         string `xml:"reason"`
		AmiLaunchIndex string `xml:"amiLaunchIndex"`
		ProductCodes   string `xml:"productCodes"`
		InstanceType   string `xml:"instanceType"`
		LaunchTime     string `xml:"launchTime"`
		Placement      struct {
			AvailabilityZone string `xml:"availabilityZone"`
			GroupName        string `xml:"groupName"`
			Tenancy          string `xml:"tenancy"`
		} `xml:"placement"`
		KernelId   string `xml:"kernelId"`
		Monitoring struct {
			State string `xml:"state"`
		} `xml:"monitoring"`
		SubnetId         string `xml:"subnetId"`
		VpcId            string `xml:"vpcId"`
		IpAddress        string `xml:"ipAddress"`
		PrivateIpAddress string `xml:"privateIpAddress"`
		SourceDestCheck  bool   `xml:"sourceDestCheck"`
		GroupSet         []struct {
			GroupId   string `xml:"groupId"`
			GroupName string `xml:"groupName"`
		} `xml:"groupSet"`
		StateReason struct {
			Code    string `xml:"code"`
			Message string `xml:"message"`
		} `xml:"stateReason"`
		Architecture        string `xml:"architecture"`
		RootDeviceType      string `xml:"rootDeviceType"`
		RootDeviceName      string `xml:"rootDeviceName"`
		BlockDeviceMapping  string `xml:"blockDeviceMapping"`
		VirtualizationType  string `xml:"virtualizationType"`
		ClientToken         string `xml:"clientToken"`
		Hypervisor          string `xml:"hypervisor"`
		NetworkInterfaceSet []struct {
			NetworkInterfaceId string `xml:"networkInterfaceId"`
			SubnetId           string `xml:"subnetId"`
			VpcId              string `xml:"vpcId"`
			Description        string `xml:"description"`
			OwnerId            string `xml:"ownerId"`
			Status             string `xml:"status"`
			MacAddress         string `xml:"macAddress"`
			PrivateIpAddress   string `xml:"privateIpAddress"`
			PrivateDnsName     string `xml:"privateDnsName"`
			SourceDestCheck    string `xml:"sourceDestCheck"`
			GroupSet           []struct {
				GroupId   string `xml:"groupId"`
				GroupName string `xml:"groupName"`
			} `xml:"groupSet>item"`
			Attachment struct {
				AttachmentId        string `xml:"attachmentId"`
				DeviceIndex         string `xml:"deviceIndex"`
				Status              string `xml:"status"`
				AttachTime          string `xml:"attachTime"`
				DeleteOnTermination bool   `xml:"deleteOnTermination"`
			} `xml:"attachment"`
			PrivateIpAddressesSet []struct {
				PrivateIpAddress string `xml:"privateIpAddress"`
				PrivateDnsName   string `xml:"privateDnsName"`
				Primary          bool   `xml:"primary"`
			} `xml:"privateIpAddressesSet>item"`
		} `xml:"networkInterfaceSet>item"`
		EbsOptimized bool `xml:"ebsOptimized"`
	}

	RunInstancesResponse struct {
		RequestId     string        `xml:"requestId"`
		ReservationId string        `xml:"reservationId"`
		OwnerId       string        `xml:"ownerId"`
		Instances     []EC2Instance `xml:"instancesSet>item"`
	}
)

func newAwsApiResponseError(r http.Response) error {
	var errorResponse ErrorResponse
	if err := getDecodedResponse(r, &errorResponse); err != nil {
		return fmt.Errorf("Error decoding error response: %s", err)
	}
	msg := ""
	for _, e := range errorResponse.Errors {
		msg += fmt.Sprintf("%s\n", e.Message)
	}
	return fmt.Errorf("Non-200 API response: code=%d message=%s", r.StatusCode, msg)
}

func newAwsApiCallError(err error) error {
	return fmt.Errorf("Problem with AWS API call: %s", err)
}

func getDecodedResponse(r http.Response, into interface{}) error {
	defer r.Body.Close()
	if err := xml.NewDecoder(r.Body).Decode(into); err != nil {
		return fmt.Errorf("Error decoding error response: %s", err)
	}
	return nil
}

func NewEC2(auth Auth, region string) *EC2 {
	endpoint := fmt.Sprintf("https://ec2.%s.amazonaws.com", region)
	return &EC2{
		Endpoint: endpoint,
		Auth:     auth,
		Region:   region,
	}
}

func (e *EC2) awsApiCall(v url.Values) (http.Response, error) {
	v.Set("Version", "2014-06-15")
	client := &http.Client{}
	finalEndpoint := fmt.Sprintf("%s?%s", e.Endpoint, v.Encode())
	req, err := http.NewRequest("GET", finalEndpoint, nil)
	if err != nil {
		return http.Response{}, fmt.Errorf("error creating request from client")
	}
	req.Header.Add("Content-type", "application/json")

	awsauth.Sign4(req, awsauth.Credentials{
		AccessKeyID:     e.Auth.AccessKey,
		SecretAccessKey: e.Auth.SecretKey,
		SecurityToken:   e.Auth.SessionToken,
	})
	resp, err := client.Do(req)
	if err != nil {
		return *resp, fmt.Errorf("client encountered error while doing the request: %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return *resp, newAwsApiResponseError(*resp)
	}
	return *resp, nil
}

func (e *EC2) RunInstance(amiId string, instanceType string, zone string, minCount int, maxCount int, securityGroup string, keyName string, subnetId string, bdm *BlockDeviceMapping) (EC2Instance, error) {
	instance := Instance{}
	v := url.Values{}
	v.Set("Action", "RunInstances")
	v.Set("ImageId", amiId)
	v.Set("Placement.AvailabilityZone", e.Region+zone)
	v.Set("MinCount", strconv.Itoa(minCount))
	v.Set("MaxCount", strconv.Itoa(maxCount))
	v.Set("KeyName", keyName)
	v.Set("InstanceType", instanceType)
	v.Set("NetworkInterface.0.DeviceIndex", "0")
	v.Set("NetworkInterface.0.SecurityGroupId.0", securityGroup)
	v.Set("NetworkInterface.0.SubnetId", subnetId)
	v.Set("NetworkInterface.0.AssociatePublicIpAddress", "1")

	if bdm != nil {
		v.Set("BlockDeviceMapping.0.DeviceName", bdm.DeviceName)
		v.Set("BlockDeviceMapping.0.VirtualName", bdm.VirtualName)
		v.Set("BlockDeviceMapping.0.Ebs.VolumeSize", strconv.FormatInt(bdm.VolumeSize, 10))
		v.Set("BlockDeviceMapping.0.Ebs.VolumeType", bdm.VolumeType)
		deleteOnTerm := 0
		if bdm.DeleteOnTermination {
			deleteOnTerm = 1
		}
		v.Set("BlockDeviceMapping.0.Ebs.DeleteOnTermination", strconv.Itoa(deleteOnTerm))
	}

	resp, err := e.awsApiCall(v)

	if err != nil {
		return instance.info, newAwsApiCallError(err)
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return instance.info, fmt.Errorf("Error reading AWS response body")
	}
	unmarshalledResponse := RunInstancesResponse{}
	err = xml.Unmarshal(contents, &unmarshalledResponse)
	if err != nil {
		return instance.info, fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	instance.info = unmarshalledResponse.Instances[0]
	return instance.info, nil
}

func (e *EC2) DeleteKeyPair(name string) error {
	v := url.Values{}
	v.Set("Action", "DeleteKeyPair")
	v.Set("KeyName", name)

	_, err := e.awsApiCall(v)
	if err != nil {
		return fmt.Errorf("Error making API call to delete keypair :%s", err)
	}
	return nil
}

func (e *EC2) CreateKeyPair(name string) ([]byte, error) {
	v := url.Values{}
	v.Set("Action", "CreateKeyPair")
	v.Set("KeyName", name)
	resp, err := e.awsApiCall(v)
	if err != nil {
		return nil, fmt.Errorf("Error trying API call to create keypair: %s", err)
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading AWS response body")
	}

	unmarshalledResponse := CreateKeyPairResponse{}
	if xml.Unmarshal(contents, &unmarshalledResponse); err != nil {
		return nil, fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	key := unmarshalledResponse.KeyMaterial

	return key, nil
}

func (e *EC2) ImportKeyPair(name, publicKey string) error {
	keyMaterial := base64.StdEncoding.EncodeToString([]byte(publicKey))

	v := url.Values{}
	v.Set("Action", "ImportKeyPair")
	v.Set("KeyName", name)
	v.Set("PublicKeyMaterial", keyMaterial)

	resp, err := e.awsApiCall(v)
	if err != nil {
		return fmt.Errorf("Error trying API call to create keypair: %s", err)
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Error reading AWS response body")
	}

	unmarshalledResponse := ImportKeyPairResponse{}
	if xml.Unmarshal(contents, &unmarshalledResponse); err != nil {
		return fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	return nil
}

func (e *EC2) CreateTags(id string, tags map[string]string) error {
	v := url.Values{}
	v.Set("Action", "CreateTags")
	v.Set("ResourceId.1", id)

	counter := 1
	for k, val := range tags {
		v.Set(fmt.Sprintf("Tag.%d.Key", counter), k)
		v.Set(fmt.Sprintf("Tag.%d.Value", counter), val)

		counter += 1
	}

	resp, err := e.awsApiCall(v)
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	createTagsResponse := &CreateTagsResponse{}

	if err := getDecodedResponse(resp, &createTagsResponse); err != nil {
		return fmt.Errorf("Error decoding create tags response: %s", err)
	}

	return nil
}

func (e *EC2) CreateSecurityGroup(name string, description string, vpcId string) (*SecurityGroup, error) {
	v := url.Values{}
	v.Set("Action", "CreateSecurityGroup")
	v.Set("GroupName", name)
	v.Set("GroupDescription", url.QueryEscape(description))
	v.Set("VpcId", vpcId)

	resp, err := e.awsApiCall(v)
	defer resp.Body.Close()
	if err != nil {
		// ugly hack since API has no way to check if SG already exists
		if resp.StatusCode == http.StatusBadRequest {
			var errorResponse ErrorResponse
			if err := getDecodedResponse(resp, &errorResponse); err != nil {
				return nil, fmt.Errorf("Error decoding error response: %s", err)
			}
			if errorResponse.Errors[0].Code == ErrorDuplicateGroup {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("Error making API call to create security group: %s", err)
	}

	createSecurityGroupResponse := CreateSecurityGroupResponse{}

	if err := getDecodedResponse(resp, &createSecurityGroupResponse); err != nil {
		return nil, fmt.Errorf("Error decoding create security groups response: %s", err)
	}

	group := &SecurityGroup{
		GroupId: createSecurityGroupResponse.GroupId,
		VpcId:   vpcId,
	}
	return group, nil
}

func (e *EC2) AuthorizeSecurityGroup(groupId string, permissions []IpPermission) error {
	v := url.Values{}
	v.Set("Action", "AuthorizeSecurityGroupIngress")
	v.Set("GroupId", groupId)

	for index, perm := range permissions {
		n := index + 1 // amazon starts counting from 1 not 0
		v.Set(fmt.Sprintf("IpPermissions.%d.IpProtocol", n), perm.Protocol)
		v.Set(fmt.Sprintf("IpPermissions.%d.FromPort", n), strconv.Itoa(perm.FromPort))
		v.Set(fmt.Sprintf("IpPermissions.%d.ToPort", n), strconv.Itoa(perm.ToPort))
		v.Set(fmt.Sprintf("IpPermissions.%d.IpRanges.1.CidrIp", n), perm.IpRange)
	}
	resp, err := e.awsApiCall(v)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Error making API call to authorize security group ingress: %s", err)
	}
	return nil
}

func (e *EC2) DeleteSecurityGroup(groupId string) error {
	v := url.Values{}
	v.Set("Action", "DeleteSecurityGroup")
	v.Set("GroupId", groupId)

	resp, err := e.awsApiCall(v)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("Error making API call to delete security group: %s", err)
	}

	deleteSecurityGroupResponse := DeleteSecurityGroupResponse{}

	if err := getDecodedResponse(resp, &deleteSecurityGroupResponse); err != nil {
		return fmt.Errorf("Error decoding delete security groups response: %s", err)
	}

	return nil
}

func (e *EC2) GetSubnets() ([]Subnet, error) {
	subnets := []Subnet{}
	resp, err := e.performStandardAction("DescribeSubnets")
	if err != nil {
		return subnets, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return subnets, fmt.Errorf("Error reading AWS response body: %s", err)
	}

	unmarshalledResponse := DescribeSubnetsResponse{}
	if err = xml.Unmarshal(contents, &unmarshalledResponse); err != nil {
		return subnets, fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	subnets = unmarshalledResponse.SubnetSet

	return subnets, nil
}
func (e *EC2) GetInstanceState(instanceId string) (state.State, error) {
	resp, err := e.performInstanceAction(instanceId, "DescribeInstances", nil)
	if err != nil {
		return state.Error, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return state.Error, fmt.Errorf("Error reading AWS response body: %s", err)
	}

	unmarshalledResponse := DescribeInstancesResponse{}
	if err = xml.Unmarshal(contents, &unmarshalledResponse); err != nil {
		return state.Error, fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	reservationSet := unmarshalledResponse.ReservationSet[0]
	instanceState := reservationSet.InstancesSet[0].InstanceState

	shortState := strings.TrimSpace(instanceState.Name)
	switch shortState {
	case "pending":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "stopped":
		return state.Stopped, nil
	case "stopping":
		return state.Stopped, nil
	}

	return state.Error, nil
}

func (e *EC2) GetInstance(instanceId string) (EC2Instance, error) {
	ec2Instance := EC2Instance{}
	resp, err := e.performInstanceAction(instanceId, "DescribeInstances", nil)
	if err != nil {
		return ec2Instance, err
	}
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ec2Instance, fmt.Errorf("Error reading AWS response body: %s", err)
	}

	unmarshalledResponse := DescribeInstancesResponse{}
	if err = xml.Unmarshal(contents, &unmarshalledResponse); err != nil {
		return ec2Instance, fmt.Errorf("Error unmarshalling AWS response XML: %s", err)
	}

	reservationSet := unmarshalledResponse.ReservationSet[0]
	instance := reservationSet.InstancesSet[0]
	return instance, nil
}

func (e *EC2) StartInstance(instanceId string) error {
	if _, err := e.performInstanceAction(instanceId, "StartInstances", nil); err != nil {
		return err
	}
	return nil
}

func (e *EC2) RestartInstance(instanceId string) error {
	if _, err := e.performInstanceAction(instanceId, "RebootInstances", nil); err != nil {
		return err
	}
	return nil
}

func (e *EC2) StopInstance(instanceId string, force bool) error {
	vars := make(map[string]string)
	if force {
		vars["Force"] = "1"
	}

	if _, err := e.performInstanceAction(instanceId, "StopInstances", &vars); err != nil {
		return err
	}
	return nil
}

func (e *EC2) TerminateInstance(instanceId string) error {
	if _, err := e.performInstanceAction(instanceId, "TerminateInstances", nil); err != nil {
		return err
	}
	return nil
}

func (e *EC2) performStandardAction(action string) (http.Response, error) {
	v := url.Values{}
	v.Set("Action", action)
	resp, err := e.awsApiCall(v)
	if err != nil {
		return resp, newAwsApiCallError(err)
	}
	return resp, nil
}

func (e *EC2) performInstanceAction(instanceId, action string, extraVars *map[string]string) (http.Response, error) {
	v := url.Values{}
	v.Set("Action", action)
	v.Set("InstanceId.1", instanceId)
	if extraVars != nil {
		for k, val := range *extraVars {
			v.Set(k, val)
		}
	}
	resp, err := e.awsApiCall(v)
	if err != nil {
		return resp, newAwsApiCallError(err)
	}
	return resp, nil
}
