package storageServiceClient

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	azure "github.com/MSOpenTech/azure-sdk-for-go"
	"strings"
)

const (
	azureXmlns                 = "http://schemas.microsoft.com/windowsazure"
	azureStorageServiceListURL = "services/storageservices"
	azureStorageServiceURL     = "services/storageservices/%s"

	blobEndpointNotFoundError = "Blob endpoint was not found in storage serice %s"
)

func GetStorageServiceList() (*StorageServiceList, error) {
	storageServiceList := new(StorageServiceList)

	response, err := azure.SendAzureGetRequest(azureStorageServiceListURL)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(response, storageServiceList)
	if err != nil {
		return storageServiceList, err
	}

	return storageServiceList, nil
}

func GetStorageServiceByName(serviceName string) (*StorageService, error) {
	if len(serviceName) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "serviceName")
	}

	storageService := new(StorageService)
	requestURL := fmt.Sprintf(azureStorageServiceURL, serviceName)
	response, err := azure.SendAzureGetRequest(requestURL)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(response, storageService)
	if err != nil {
		return nil, err
	}

	return storageService, nil
}

func GetStorageServiceByLocation(location string) (*StorageService, error) {
	if len(location) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "location")
	}

	storageService := new(StorageService)
	storageServiceList, err := GetStorageServiceList()
	if err != nil {
		return storageService, err
	}

	for _, storageService := range storageServiceList.StorageServices {
		if storageService.StorageServiceProperties.Location != location {
			continue
		}

		return &storageService, nil
	}

	return nil, nil
}

func CreateStorageService(name, location string) (*StorageService, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "name")
	}
	if len(location) == 0 {
		return nil, fmt.Errorf(azure.ParamNotSpecifiedError, "location")
	}

	storageDeploymentConfig := createStorageServiceDeploymentConf(name, location)
	deploymentBytes, err := xml.Marshal(storageDeploymentConfig)
	if err != nil {
		return nil, err
	}

	requestId, err := azure.SendAzurePostRequest(azureStorageServiceListURL, deploymentBytes)
	if err != nil {
		return nil, err
	}

	azure.WaitAsyncOperation(requestId)
	storageService, err := GetStorageServiceByName(storageDeploymentConfig.ServiceName)
	if err != nil {
		return nil, err
	}

	return storageService, nil
}

func GetBlobEndpoint(storageService *StorageService) (string, error) {
	for _, endpoint := range storageService.StorageServiceProperties.Endpoints {
		if !strings.Contains(endpoint, ".blob.core") {
			continue
		}

		return endpoint, nil
	}

	return "", errors.New(fmt.Sprintf(blobEndpointNotFoundError, storageService.ServiceName))
}

func createStorageServiceDeploymentConf(name, location string) StorageServiceDeployment {
	storageServiceDeployment := StorageServiceDeployment{}

	storageServiceDeployment.ServiceName = name
	label := base64.StdEncoding.EncodeToString([]byte(name))
	storageServiceDeployment.Label = label
	storageServiceDeployment.Location = location
	storageServiceDeployment.Xmlns = azureXmlns

	return storageServiceDeployment
}
