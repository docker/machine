package vmDiskClient

import (
	"fmt"
	azure "github.com/MSOpenTech/azure-sdk-for-go"
)

const (
	azureVMDiskURL = "services/disks/%s"
)

//Region public methods starts

func DeleteDisk(diskName string) error {
	if len(diskName) == 0 {
		return fmt.Errorf(azure.ParamNotSpecifiedError, "diskName")
	}

	requestURL := fmt.Sprintf(azureVMDiskURL, diskName)
	requestId, err := azure.SendAzureDeleteRequest(requestURL)
	if err != nil {
		return err
	}

	azure.WaitAsyncOperation(requestId)
	return nil
}

//Region public methods ends
