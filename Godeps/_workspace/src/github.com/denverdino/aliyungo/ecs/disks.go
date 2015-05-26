package ecs

import (
	"github.com/denverdino/aliyungo/util"
	"time"
)

// Types of disks
type DiskType string

const (
	DiskTypeAll       = DiskType("all") //Default
	DiskTypeAllSystem = DiskType("system")
	DiskTypeAllData   = DiskType("data")
)

// Categories of disks
type DiskCategory string

const (
	DiskCategoryAll          = DiskCategory("all") //Default
	DiskCategoryCloud        = DiskCategory("cloud")
	DiskCategoryEphemeral    = DiskCategory("ephemeral")
	DiskCategoryEphemeralSSD = DiskCategory("ephemeral_ssd")
)

// Status of disks
type DiskStatus string

const (
	DiskStatusInUse     = DiskStatus("In_use")
	DiskStatusAvailable = DiskStatus("Available")
	DiskStatusAttaching = DiskStatus("Attaching")
	DiskStatusDetaching = DiskStatus("Detaching")
	DiskStatusCreating  = DiskStatus("Creating")
	DiskStatusReIniting = DiskStatus("ReIniting")
	DiskStatusAll       = DiskStatus("All") //Default
)

// A DescribeDisksArgs defines the arguments to describe disks
type DescribeDisksArgs struct {
	RegionId           Region
	ZoneId             string
	DiskIds            []string
	InstanceId         string
	DiskType           DiskType     //enum for all(default) | system | data
	Category           DiskCategory //enum for all(default) | cloud | ephemeral
	Status             DiskStatus   //enum for In_use | Available | Attaching | Detaching | Creating | ReIniting | All(default)
	SnapshotId         string
	Name               string
	Portable           *bool //optional
	DeleteWithInstance *bool //optional
	DeleteAutoSnapshot *bool //optional
	Pagination
}

type DiskItemType struct {
	DiskId             string
	RegionId           Region
	ZoneId             string
	DiskName           string
	Description        string
	Type               DiskType
	Category           DiskCategory
	Size               int
	ImageId            string
	SourceSnapshotId   string
	ProductCode        string
	Portable           bool
	Status             DiskStatus
	OperationLocks     OperationLocksType
	InstanceId         string
	Device             string
	DeleteWithInstance bool
	DeleteAutoSnapshot bool
	EnableAutoSnapshot bool
	CreationTime       util.ISO6801Time
	AttachedTime       util.ISO6801Time
	DetachedTime       util.ISO6801Time
}

type DescribeDisksResponse struct {
	CommonResponse

	RegionId Region
	PaginationResult
	Disks struct {
		Disk []DiskItemType
	}
}

// DescribeDisks describes Disks
func (client *Client) DescribeDisks(args *DescribeDisksArgs) (disks []DiskItemType, pagination *PaginationResult, err error) {
	response := DescribeDisksResponse{}

	err = client.Invoke("DescribeDisks", args, &response)

	if err != nil {
		return nil, nil, err
	}

	return response.Disks.Disk, &response.PaginationResult, err
}

type CreateDiskArgs struct {
	RegionId    Region
	ZoneId      string
	DiskName    string
	Description string
	Size        int
	SnapshotId  string
	ClientToken string
}

type CreateDisksResponse struct {
	CommonResponse
	DiskId string
}

// CreateDisk creates a new disk
func (client *Client) CreateDisk(args *CreateDiskArgs) (diskId string, err error) {
	response := CreateDisksResponse{}
	err = client.Invoke("CreateDisk", args, &response)
	if err != nil {
		return "", err
	}
	return response.DiskId, err
}

type DeleteDiskArgs struct {
	DiskId string
}

type DeleteDiskResponse struct {
	CommonResponse
}

// DeleteDisk deletes disk
func (client *Client) DeleteDisk(diskId string) error {
	args := DeleteDiskArgs{
		DiskId: diskId,
	}
	response := DeleteDiskResponse{}
	err := client.Invoke("DeleteDisk", &args, &response)
	return err
}

type ReInitDiskArgs struct {
	DiskId string
}

type ReInitDiskResponse struct {
	CommonResponse
}

// ReInitDisk reinitizes disk
func (client *Client) ReInitDisk(diskId string) error {
	args := ReInitDiskArgs{
		DiskId: diskId,
	}
	response := ReInitDiskResponse{}
	err := client.Invoke("ReInitDisk", &args, &response)
	return err
}

type AttachDiskArgs struct {
	InstanceId         string
	DiskId             string
	Device             string
	DeleteWithInstance bool
}

type AttachDiskResponse struct {
	CommonResponse
}

// AttachDisk attaches disk to instance
func (client *Client) AttachDisk(args *AttachDiskArgs) error {
	response := AttachDiskResponse{}
	err := client.Invoke("AttachDisk", args, &response)
	return err
}

type DetachDiskArgs struct {
	InstanceId string
	DiskId     string
}

type DetachDiskResponse struct {
	CommonResponse
}

// DetachDisk detaches disk from instance
func (client *Client) DetachDisk(instanceId string, diskId string) error {
	args := DetachDiskArgs{
		InstanceId: instanceId,
		DiskId:     diskId,
	}
	response := DetachDiskResponse{}
	err := client.Invoke("DetachDisk", &args, &response)
	return err
}

type ResetDiskArgs struct {
	DiskId     string
	SnapshotId string
}

type ResetDiskResponse struct {
	CommonResponse
}

// ResetDisk resets disk to original status
func (client *Client) ResetDisk(diskId string, snapshotId string) error {
	args := ResetDiskArgs{
		SnapshotId: snapshotId,
		DiskId:     diskId,
	}
	response := ResetDiskResponse{}
	err := client.Invoke("ResetDisk", &args, &response)
	return err
}

type ModifyDiskAttributeArgs struct {
	DiskId             string
	DiskName           string
	Description        string
	DeleteWithInstance *bool
	DeleteAutoSnapshot *bool
	EnableAutoSnapshot *bool
}

type ModifyDiskAttributeResponse struct {
	CommonResponse
}

// ModifyDiskAttribute modifies disk attribute
func (client *Client) ModifyDiskAttribute(args *ModifyDiskAttributeArgs) error {
	response := ModifyDiskAttributeResponse{}
	err := client.Invoke("ModifyDiskAttribute", &args, &response)
	return err
}

// WaitForDisk waits for disk to given status
func (client *Client) WaitForDisk(regionId Region, diskId string, status DiskStatus, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	args := DescribeDisksArgs{
		RegionId: regionId,
		DiskIds:  []string{diskId},
	}

	for {
		disks, _, err := client.DescribeDisks(&args)
		if err != nil {
			return err
		}
		if disks == nil || len(disks) == 0 {
			return getECSErrorFromString("Not found")
		}
		if disks[0].Status == status {
			break
		}
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return getECSErrorFromString("Timeout")
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
	}
	return nil
}
