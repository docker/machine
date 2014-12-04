package vmClient

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
	"unicode"

	azure "github.com/MSOpenTech/azure-sdk-for-go"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/imageClient"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/locationClient"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/storageServiceClient"
	"github.com/MSOpenTech/azure-sdk-for-go/clients/vmDiskClient"
)

const (
	azureXmlns                        = "http://schemas.microsoft.com/windowsazure"
	azureDeploymentListURL            = "services/hostedservices/%s/deployments"
	azureHostedServiceListURL         = "services/hostedservices"
	azureHostedServiceURL             = "services/hostedservices/%s"
	azureHostedServiceAvailabilityURL = "services/hostedservices/operations/isavailable/%s"
	azureDeploymentURL                = "services/hostedservices/%s/deployments/%s"
	azureRoleURL                      = "services/hostedservices/%s/deployments/%s/roles/%s"
	azureOperationsURL                = "services/hostedservices/%s/deployments/%s/roleinstances/%s/Operations"
	azureCertificatListURL            = "services/hostedservices/%s/certificates"
	azureRoleSizeListURL              = "rolesizes"

	osLinux                   = "Linux"
	osWindows                 = "Windows"
	dockerPublicConfigVersion = 2

	provisioningConfDoesNotExistsError = "You should set azure VM provisioning config first"
	invalidCertExtensionError          = "Certificate %s is invalid. Please specify %s certificate."
	invalidOSError                     = "You must specify correct OS param. Valid values are 'Linux' and 'Windows'"
	invalidDnsLengthError              = "The DNS name must be between 3 and 25 characters."
	invalidPasswordLengthError         = "Password must be between 4 and 30 characters."
	invalidPasswordError               = "Password must have at least one upper case, lower case and numeric character."
	invalidRoleSizeError               = "Invalid role size: %s. Available role sizes: %s"
)

//Region public methods starts

func CreateAzureVM(azureVMConfiguration *Role, dnsName, location string) error {
	if azureVMConfiguration == nil {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "azureVMConfiguration")
	}
	if len(dnsName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "dnsName")
	}
	if len(location) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "location")
	}

	err := verifyDNSname(dnsName)
	if err != nil {
		return err
	}

	requestId, err := CreateHostedService(dnsName, location)
	if err != nil {
		return err
	}

	azure.WaitAsyncOperation(requestId)

	if azureVMConfiguration.UseCertAuth {
		err = uploadServiceCert(dnsName, azureVMConfiguration.CertPath)
		if err != nil {
			DeleteHostedService(dnsName)
			return err
		}
	}

	vMDeployment := createVMDeploymentConfig(azureVMConfiguration)
	vMDeploymentBytes, err := xml.Marshal(vMDeployment)
	if err != nil {
		DeleteHostedService(dnsName)
		return err
	}

	requestURL := fmt.Sprintf(azureDeploymentListURL, azureVMConfiguration.RoleName)
	requestId, err = azure.SendAzurePostRequest(requestURL, vMDeploymentBytes)
	if err != nil {
		DeleteHostedService(dnsName)
		return err
	}

	azure.WaitAsyncOperation(requestId)

	return nil
}

func CreateHostedService(dnsName, location string) (string, error) {
	if len(dnsName) == 0 {
		return "", fmt.Errorf(azure.ParamNotSpecifiedError, "dnsName")
	}
	if len(location) == 0 {
		return "", fmt.Errorf(azure.ParamNotSpecifiedError, "location")
	}

	err := verifyDNSname(dnsName)
	if err != nil {
		return "", err
	}

	result, reason, err := CheckHostedServiceNameAvailability(dnsName)
	if err != nil {
		return "", err
	}
	if !result {
		return "", fmt.Errorf("%s Hosted service name: %s", reason, dnsName)
	}

	err = locationClient.ResolveLocation(location)
	if err != nil {
		return "", err
	}

	hostedServiceDeployment := createHostedServiceDeploymentConfig(dnsName, location)
	hostedServiceBytes, err := xml.Marshal(hostedServiceDeployment)
	if err != nil {
		return "", err
	}

	requestURL := azureHostedServiceListURL
	requestId, err := azure.SendAzurePostRequest(requestURL, hostedServiceBytes)
	if err != nil {
		return "", err
	}

	return requestId, nil
}

func CheckHostedServiceNameAvailability(dnsName string) (bool, string, error) {
	if len(dnsName) == 0 {
		return false, "", fmt.Errorf(azure.ParamNotSpecifiedError, "dnsName")
	}

	err := verifyDNSname(dnsName)
	if err != nil {
		return false, "", err
	}

	requestURL := fmt.Sprintf(azureHostedServiceAvailabilityURL, dnsName)
	response, err := azure.SendAzureGetRequest(requestURL)
	if err != nil {
		return false, "", err
	}

	availabilityResponse := new(AvailabilityResponse)
	err = xml.Unmarshal(response, availabilityResponse)
	if err != nil {
		return false, "", err
	}

	return availabilityResponse.Result, availabilityResponse.Reason, nil
}

func DeleteHostedService(dnsName string) error {
	if len(dnsName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "dnsName")
	}

	err := verifyDNSname(dnsName)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(azureHostedServiceURL, dnsName)
	requestId, err := azure.SendAzureDeleteRequest(requestURL)
	if err != nil {
		return err
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

func CreateAzureVMConfiguration(dnsName, instanceSize, imageName, location string) (*Role, error) {
	if len(dnsName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "dnsName")
	}
	if len(instanceSize) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "instanceSize")
	}
	if len(imageName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "imageName")
	}
	if len(location) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "location")
	}

	err := verifyDNSname(dnsName)
	if err != nil {
		return nil, err
	}

	err = locationClient.ResolveLocation(location)
	if err != nil {
		return nil, err
	}

	err = ResolveRoleSize(instanceSize)
	if err != nil {
		return nil, err
	}

	role, err := createAzureVMRole(dnsName, instanceSize, imageName, location)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func AddAzureLinuxProvisioningConfig(azureVMConfiguration *Role, userName, password, certPath string, sshPort int) (*Role, error) {
	if azureVMConfiguration == nil {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "azureVMConfiguration")
	}
	if len(userName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "userName")
	}

	configurationSets := ConfigurationSets{}
	provisioningConfig, err := createLinuxProvisioningConfig(azureVMConfiguration.RoleName, userName, password, certPath)
	if err != nil {
		return nil, err
	}

	configurationSets.ConfigurationSet = append(configurationSets.ConfigurationSet, provisioningConfig)

	networkConfig, networkErr := createNetworkConfig(osLinux, sshPort)
	if networkErr != nil {
		return nil, err
	}

	configurationSets.ConfigurationSet = append(configurationSets.ConfigurationSet, networkConfig)

	azureVMConfiguration.ConfigurationSets = configurationSets

	if len(certPath) > 0 {
		azureVMConfiguration.UseCertAuth = true
		azureVMConfiguration.CertPath = certPath
	}

	return azureVMConfiguration, nil
}

func SetAzureVMExtension(azureVMConfiguration *Role, name string, publisher string, version string, referenceName string, state string, publicConfigurationValue string, privateConfigurationValue string) (*Role, error) {
	if azureVMConfiguration == nil {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "azureVMConfiguration")
	}
	if len(name) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "name")
	}
	if len(publisher) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "publisher")
	}
	if len(version) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "version")
	}
	if len(referenceName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "referenceName")
	}

	extension := ResourceExtensionReference{}
	extension.Name = name
	extension.Publisher = publisher
	extension.Version = version
	extension.ReferenceName = referenceName
	extension.State = state

	if len(privateConfigurationValue) > 0 {
		privateConfig := ResourceExtensionParameter{}
		privateConfig.Key = "ignored"
		privateConfig.Value = base64.StdEncoding.EncodeToString([]byte(privateConfigurationValue))
		privateConfig.Type = "Private"

		extension.ResourceExtensionParameterValues.ResourceExtensionParameterValue = append(extension.ResourceExtensionParameterValues.ResourceExtensionParameterValue, privateConfig)
	}

	if len(publicConfigurationValue) > 0 {
		publicConfig := ResourceExtensionParameter{}
		publicConfig.Key = "ignored"
		publicConfig.Value = base64.StdEncoding.EncodeToString([]byte(publicConfigurationValue))
		publicConfig.Type = "Public"

		extension.ResourceExtensionParameterValues.ResourceExtensionParameterValue = append(extension.ResourceExtensionParameterValues.ResourceExtensionParameterValue, publicConfig)
	}

	azureVMConfiguration.ResourceExtensionReferences.ResourceExtensionReference = append(azureVMConfiguration.ResourceExtensionReferences.ResourceExtensionReference, extension)

	return azureVMConfiguration, nil
}

func SetAzureDockerVMExtension(azureVMConfiguration *Role, dockerPort int, version string) (*Role, error) {
	if azureVMConfiguration == nil {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "azureVMConfiguration")
	}

	if len(version) == 0 {
		version = "0.3"
	}

	err := addDockerPort(azureVMConfiguration.ConfigurationSets.ConfigurationSet, dockerPort)
	if err != nil {
		return nil, err
	}

	publicConfiguration, err := createDockerPublicConfig(dockerPort)
	if err != nil {
		return nil, err
	}

	privateConfiguration := "{}"
	if err != nil {
		return nil, err
	}

	azureVMConfiguration, err = SetAzureVMExtension(azureVMConfiguration, "DockerExtension", "MSOpenTech.Extensions", version, "DockerExtension", "enable", publicConfiguration, privateConfiguration)
	return azureVMConfiguration, nil
}

func GetVMDeployment(cloudserviceName, deploymentName string) (*VMDeployment, error) {
	if len(cloudserviceName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}

	deployment := new(VMDeployment)

	requestURL := fmt.Sprintf(azureDeploymentURL, cloudserviceName, deploymentName)
	response, azureErr := azure.SendAzureGetRequest(requestURL)
	if azureErr != nil {
		return nil, azureErr
	}

	err := xml.Unmarshal(response, deployment)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func DeleteVMDeployment(cloudserviceName, deploymentName string) error {
	if len(cloudserviceName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}

	vmDeployment, err := GetVMDeployment(cloudserviceName, deploymentName)
	if err != nil {
		return err
	}
	vmDiskName := vmDeployment.RoleList.Role[0].OSVirtualHardDisk.DiskName

	requestURL := fmt.Sprintf(azureDeploymentURL, cloudserviceName, deploymentName)
	requestId, err := azure.SendAzureDeleteRequest(requestURL)
	if err != nil {
		return err
	}

	azure.WaitAsyncOperation(requestId)

	err = vmDiskClient.DeleteDisk(vmDiskName)
	if err != nil {
		return err
	}

	return nil
}

func GetRole(cloudserviceName, deploymentName, roleName string) (*Role, error) {
	if len(cloudserviceName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}
	if len(roleName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "roleName")
	}

	role := new(Role)

	requestURL := fmt.Sprintf(azureRoleURL, cloudserviceName, deploymentName, roleName)
	response, azureErr := azure.SendAzureGetRequest(requestURL)
	if azureErr != nil {
		return nil, azureErr
	}

	err := xml.Unmarshal(response, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func StartRole(cloudserviceName, deploymentName, roleName string) error {
	if len(cloudserviceName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}
	if len(roleName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "roleName")
	}

	startRoleOperation := createStartRoleOperation()

	startRoleOperationBytes, err := xml.Marshal(startRoleOperation)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudserviceName, deploymentName, roleName)
	requestId, azureErr := azure.SendAzurePostRequest(requestURL, startRoleOperationBytes)
	if azureErr != nil {
		return azureErr
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

func ShutdownRole(cloudserviceName, deploymentName, roleName string) error {
	if len(cloudserviceName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}
	if len(roleName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "roleName")
	}

	shutdownRoleOperation := createShutdowRoleOperation()

	shutdownRoleOperationBytes, err := xml.Marshal(shutdownRoleOperation)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudserviceName, deploymentName, roleName)
	requestId, azureErr := azure.SendAzurePostRequest(requestURL, shutdownRoleOperationBytes)
	if azureErr != nil {
		return azureErr
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

func RestartRole(cloudserviceName, deploymentName, roleName string) error {
	if len(cloudserviceName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}
	if len(roleName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "roleName")
	}

	restartRoleOperation := createRestartRoleOperation()

	restartRoleOperationBytes, err := xml.Marshal(restartRoleOperation)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(azureOperationsURL, cloudserviceName, deploymentName, roleName)
	requestId, azureErr := azure.SendAzurePostRequest(requestURL, restartRoleOperationBytes)
	if azureErr != nil {
		return azureErr
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

func DeleteRole(cloudserviceName, deploymentName, roleName string) error {
	if len(cloudserviceName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "cloudserviceName")
	}
	if len(deploymentName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "deploymentName")
	}
	if len(roleName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "roleName")
	}

	requestURL := fmt.Sprintf(azureRoleURL, cloudserviceName, deploymentName, roleName)
	requestId, azureErr := azure.SendAzureDeleteRequest(requestURL)
	if azureErr != nil {
		return azureErr
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

func GetRoleSizeList() (RoleSizeList, error) {
	roleSizeList := RoleSizeList{}

	response, err := azure.SendAzureGetRequest(azureRoleSizeListURL)
	if err != nil {
		return roleSizeList, err
	}

	err = xml.Unmarshal(response, &roleSizeList)
	if err != nil {
		return roleSizeList, err
	}

	return roleSizeList, err
}

func ResolveRoleSize(roleSizeName string) error {
	if len(roleSizeName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "roleSizeName")
	}

	roleSizeList, err := GetRoleSizeList()
	if err != nil {
		return err
	}

	for _, roleSize := range roleSizeList.RoleSizes {
		if roleSize.Name != roleSizeName {
			continue
		}

		return nil
	}

	var availableSizes bytes.Buffer
	for _, existingSize := range roleSizeList.RoleSizes {
		availableSizes.WriteString(existingSize.Name + ", ")
	}

	return errors.New(fmt.Sprintf(invalidRoleSizeError, roleSizeName, strings.Trim(availableSizes.String(), ", ")))
}

//Region public methods ends

//Region private methods starts

func createStartRoleOperation() StartRoleOperation {
	startRoleOperation := StartRoleOperation{}
	startRoleOperation.OperationType = "StartRoleOperation"
	startRoleOperation.Xmlns = azureXmlns

	return startRoleOperation
}

func createShutdowRoleOperation() ShutdownRoleOperation {
	shutdownRoleOperation := ShutdownRoleOperation{}
	shutdownRoleOperation.OperationType = "ShutdownRoleOperation"
	shutdownRoleOperation.Xmlns = azureXmlns

	return shutdownRoleOperation
}

func createRestartRoleOperation() RestartRoleOperation {
	startRoleOperation := RestartRoleOperation{}
	startRoleOperation.OperationType = "RestartRoleOperation"
	startRoleOperation.Xmlns = azureXmlns

	return startRoleOperation
}

func createDockerPublicConfig(dockerPort int) (string, error) {
	config := dockerPublicConfig{DockerPort: dockerPort, Version: dockerPublicConfigVersion}
	configJson, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	return string(configJson), nil
}

func addDockerPort(configurationSets []ConfigurationSet, dockerPort int) error {
	if len(configurationSets) == 0 {
		return errors.New(provisioningConfDoesNotExistsError)
	}

	for i := 0; i < len(configurationSets); i++ {
		if configurationSets[i].ConfigurationSetType != "NetworkConfiguration" {
			continue
		}

		dockerEndpoint := createEndpoint("docker", "tcp", dockerPort, dockerPort)
		configurationSets[i].InputEndpoints.InputEndpoint = append(configurationSets[i].InputEndpoints.InputEndpoint, dockerEndpoint)
	}

	return nil
}

func createHostedServiceDeploymentConfig(dnsName, location string) HostedServiceDeployment {
	deployment := HostedServiceDeployment{}
	deployment.ServiceName = dnsName
	label := base64.StdEncoding.EncodeToString([]byte(dnsName))
	deployment.Label = label
	deployment.Location = location
	deployment.Xmlns = azureXmlns

	return deployment
}

func createVMDeploymentConfig(role *Role) VMDeployment {
	deployment := VMDeployment{}
	deployment.Name = role.RoleName
	deployment.Xmlns = azureXmlns
	deployment.DeploymentSlot = "Production"
	deployment.Label = role.RoleName
	deployment.RoleList.Role = append(deployment.RoleList.Role, role)

	return deployment
}

func createAzureVMRole(name, instanceSize, imageName, location string) (*Role, error) {
	config := new(Role)
	config.RoleName = name
	config.RoleSize = instanceSize
	config.RoleType = "PersistentVMRole"
	config.ProvisionGuestAgent = true
	var err error
	config.OSVirtualHardDisk, err = createOSVirtualHardDisk(name, imageName, location)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func createOSVirtualHardDisk(dnsName, imageName, location string) (OSVirtualHardDisk, error) {
	oSVirtualHardDisk := OSVirtualHardDisk{}

	err := imageClient.ResolveImageName(imageName)
	if err != nil {
		return oSVirtualHardDisk, err
	}

	oSVirtualHardDisk.SourceImageName = imageName
	oSVirtualHardDisk.MediaLink, err = getVHDMediaLink(dnsName, location)
	if err != nil {
		return oSVirtualHardDisk, err
	}

	return oSVirtualHardDisk, nil
}

func getVHDMediaLink(dnsName, location string) (string, error) {

	storageService, err := storageServiceClient.GetStorageServiceByLocation(location)
	if err != nil {
		return "", err
	}

	if storageService == nil {

		uuid, err := azure.NewUUID()
		if err != nil {
			return "", err
		}

		serviceName := "portalvhds" + uuid
		storageService, err = storageServiceClient.CreateStorageService(serviceName, location)
		if err != nil {
			return "", err
		}
	}

	blobEndpoint, err := storageServiceClient.GetBlobEndpoint(storageService)
	if err != nil {
		return "", err
	}

	vhdMediaLink := blobEndpoint + "vhds/" + dnsName + "-" + time.Now().Local().Format("20060102150405") + ".vhd"
	return vhdMediaLink, nil
}

func createLinuxProvisioningConfig(dnsName, userName, userPassword, certPath string) (ConfigurationSet, error) {
	provisioningConfig := ConfigurationSet{}

	disableSshPasswordAuthentication := false
	if len(userPassword) == 0 {
		disableSshPasswordAuthentication = true
		// We need to set dummy password otherwise azure API will throw an error
		userPassword = "P@ssword1"
	} else {
		err := verifyPassword(userPassword)
		if err != nil {
			return provisioningConfig, err
		}
	}

	provisioningConfig.DisableSshPasswordAuthentication = disableSshPasswordAuthentication
	provisioningConfig.ConfigurationSetType = "LinuxProvisioningConfiguration"
	provisioningConfig.HostName = dnsName
	provisioningConfig.UserName = userName
	provisioningConfig.UserPassword = userPassword

	if len(certPath) > 0 {
		var err error
		provisioningConfig.SSH, err = createSshConfig(certPath, userName)
		if err != nil {
			return provisioningConfig, err
		}
	}

	return provisioningConfig, nil
}

func uploadServiceCert(dnsName, certPath string) error {
	certificateConfig, err := createServiceCertDeploymentConf(certPath)
	if err != nil {
		return err
	}

	certificateConfigBytes, err := xml.Marshal(certificateConfig)
	if err != nil {
		return err
	}

	requestURL := fmt.Sprintf(azureCertificatListURL, dnsName)
	requestId, azureErr := azure.SendAzurePostRequest(requestURL, certificateConfigBytes)
	if azureErr != nil {
		return azureErr
	}

	err = azure.WaitAsyncOperation(requestId)
	return err
}

func createServiceCertDeploymentConf(certPath string) (ServiceCertificate, error) {
	certConfig := ServiceCertificate{}
	certConfig.Xmlns = azureXmlns
	data, err := ioutil.ReadFile(certPath)
	if err != nil {
		return certConfig, err
	}

	certData := base64.StdEncoding.EncodeToString(data)
	certConfig.Data = certData
	certConfig.CertificateFormat = "pfx"

	return certConfig, nil
}

func createSshConfig(certPath, userName string) (SSH, error) {
	sshConfig := SSH{}
	publicKey := PublicKey{}

	err := checkServiceCertExtension(certPath)
	if err != nil {
		return sshConfig, err
	}

	fingerprint, err := getServiceCertFingerprint(certPath)
	if err != nil {
		return sshConfig, err
	}

	publicKey.Fingerprint = fingerprint
	publicKey.Path = "/home/" + userName + "/.ssh/authorized_keys"

	sshConfig.PublicKeys.PublicKey = append(sshConfig.PublicKeys.PublicKey, publicKey)
	return sshConfig, nil
}

func getServiceCertFingerprint(certPath string) (string, error) {
	certData, readErr := ioutil.ReadFile(certPath)
	if readErr != nil {
		return "", readErr
	}

	block, rest := pem.Decode(certData)
	if block == nil {
		return "", errors.New(string(rest))
	}

	sha1sum := sha1.Sum(block.Bytes)
	fingerprint := fmt.Sprintf("%X", sha1sum)
	return fingerprint, nil
}

func checkServiceCertExtension(certPath string) error {
	certParts := strings.Split(certPath, ".")
	certExt := certParts[len(certParts)-1]

	acceptedExtension := "pem"
	if certExt != acceptedExtension {
		return errors.New(fmt.Sprintf(invalidCertExtensionError, certPath, acceptedExtension))
	}

	return nil
}

func createNetworkConfig(os string, sshPort int) (ConfigurationSet, error) {
	networkConfig := ConfigurationSet{}
	networkConfig.ConfigurationSetType = "NetworkConfiguration"

	var endpoint InputEndpoint
	if os == osLinux {
		endpoint = createEndpoint("ssh", "tcp", sshPort, 22)
	} else if os == osWindows {
		//!TODO add rdp endpoint
	} else {
		return networkConfig, errors.New(fmt.Sprintf(invalidOSError))
	}

	networkConfig.InputEndpoints.InputEndpoint = append(networkConfig.InputEndpoints.InputEndpoint, endpoint)

	return networkConfig, nil
}

func createEndpoint(name string, protocol string, extertalPort int, internalPort int) InputEndpoint {
	endpoint := InputEndpoint{}
	endpoint.Name = name
	endpoint.Protocol = protocol
	endpoint.Port = extertalPort
	endpoint.LocalPort = internalPort

	return endpoint
}

func verifyDNSname(dns string) error {
	if len(dns) < 3 || len(dns) > 25 {
		return fmt.Errorf(invalidDnsLengthError)
	}

	return nil
}

func verifyPassword(password string) error {
	if len(password) < 4 || len(password) > 30 {
		return fmt.Errorf(invalidPasswordLengthError)
	}

next:
	for _, classes := range map[string][]*unicode.RangeTable{
		"upper case": {unicode.Upper, unicode.Title},
		"lower case": {unicode.Lower},
		"numeric":    {unicode.Number, unicode.Digit},
	} {
		for _, r := range password {
			if unicode.IsOneOf(classes, r) {
				continue next
			}
		}
		return fmt.Errorf(invalidPasswordError)
	}
	return nil
}

//Region private methods ends
