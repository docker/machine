package ecs

import (
	"github.com/denverdino/aliyungo/util"
)

// ImageOwnerAlias represents image owner
type ImageOwnerAlias string

// Constants of image owner
const (
	ImageOwnerSystem      = ImageOwnerAlias("system")
	ImageOwnerSelf        = ImageOwnerAlias("self")
	ImageOwnerOthers      = ImageOwnerAlias("others")
	ImageOwnerMarketplace = ImageOwnerAlias("marketplace")
	ImageOwnerDefault     = ImageOwnerAlias("") //Return the values for system, self, and others
)

// DescribeImagesArgs repsents arguements to describe images
type DescribeImagesArgs struct {
	RegionId        Region
	ImageId         string
	SnapshotId      string
	ImageName       string
	ImageOwnerAlias ImageOwnerAlias
	Pagination
}

type DescribeImagesResponse struct {
	CommonResponse

	Pagination string
	PaginationResult
	Images struct {
		Image []ImageType
	}
}

type DiskDeviceMapping struct {
	SnapshotId string
	//Why Size Field is string-type.
	Size   string
	Device string
}

type ImageStatus string

const (
	ImageStatusAvailable    = ImageStatus("Available")
	ImageStatusUnAvailable  = ImageStatus("UnAvailable")
	ImageStatusCreating     = ImageStatus("Creating")
	ImageStatusCreateFailed = ImageStatus("CreateFailed")
)

type ImageType struct {
	ImageId            string
	ImageVersion       string
	Architecture       string
	ImageName          string
	Description        string
	Size               int
	ImageOwnerAlias    string
	OSName             string
	DiskDeviceMappings struct {
		DiskDeviceMapping []DiskDeviceMapping
	}
	ProductCode  string
	IsSubscribed bool
	Progress     string
	Status       ImageStatus
	CreationTime util.ISO6801Time
}

// DescribeImages describes images
func (client *Client) DescribeImages(args *DescribeImagesArgs) (images []ImageType, pagination *PaginationResult, err error) {

	args.validate()
	response := DescribeImagesResponse{}
	err = client.Invoke("DescribeImages", args, &response)
	if err != nil {
		return nil, nil, err
	}
	return response.Images.Image, &response.PaginationResult, nil
}

// CreateImageArgs repsents arguements to create image
type CreateImageArgs struct {
	RegionId     Region
	SnapshotId   string
	ImageName    string
	ImageVersion string
	Description  string
	ClientToken  string
}

type CreateImageResponse struct {
	CommonResponse

	ImageId string
}

// CreateImage creates a new image
func (client *Client) CreateImage(args *CreateImageArgs) (imageId string, err error) {
	response := &CreateImageResponse{}
	err = client.Invoke("CreateImage", args, &response)
	if err != nil {
		return "", err
	}
	return response.ImageId, nil
}

type DeleteImageArgs struct {
	RegionId Region
	ImageId  string
}

type DeleteImageResponse struct {
	CommonResponse
}

// DeleteImage deletes Image
func (client *Client) DeleteImage(regionId Region, imageId string) error {
	args := DeleteImageArgs{
		RegionId: regionId,
		ImageId:  imageId,
	}

	response := &DeleteImageResponse{}
	return client.Invoke("DeleteImage", &args, &response)
}
