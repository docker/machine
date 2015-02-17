package linodego

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Linode Disk Service
type LinodeDiskService struct {
	client *Client
}

// Response for linode.disk.list API
type LinodeDiskListResponse struct {
	Response
	Disks []Disk
}

// Response for general Disk jobs APIs
type LinodeDiskJobResponse struct {
	Response
	DiskJob DiskJob
}

// List all disks. If diskId is greater than 0, limit the results to given disk.
func (t *LinodeDiskService) List(linodeId int, diskId int) (*LinodeDiskListResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if diskId > 0 {
		u.Add("DiskID", strconv.Itoa(diskId))
	}
	v := LinodeDiskListResponse{}
	if err := t.client.do("linode.disk.list", u, &v.Response); err != nil {
		return nil, err
	}

	v.Disks = make([]Disk, 5)
	if err := json.Unmarshal(v.RawData, &v.Disks); err != nil {
		return nil, err
	}
	return &v, nil
}

// Create disk
func (t *LinodeDiskService) Create(linodeId int, diskType string, label string, size int, args map[string]string) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("Size", strconv.Itoa(size))
	u.Add("Type", diskType)
	u.Add("Label", label)
	// add optional parameters
	for k, v := range args {
		u.Add(k, v)
	}
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.create", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Create from Distribution
func (t *LinodeDiskService) CreateFromDistribution(distributionId int, linodeId int, label string, size int, args map[string]string) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DistributionID", strconv.Itoa(distributionId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("Size", strconv.Itoa(size))
	u.Add("Label", label)
	// add optional parameters
	for k, v := range args {
		u.Add(k, v)
	}
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.createFromDistribution", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Create from image
func (t *LinodeDiskService) CreateFromImage(imageId int, linodeId int, label string, size int, args map[string]string) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("ImageID", strconv.Itoa(imageId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if size > 0 {
		u.Add("Size", strconv.Itoa(size))
	}
	u.Add("Label", label)
	// add optional parameters
	for k, v := range args {
		u.Add(k, v)
	}
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.createfromimage", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Create from stackscript
func (t *LinodeDiskService) CreateFromStackscript(
	stackScriptId int, linodeId int, label string,
	stackScriptUDFResponses string,
	distributionId int, size int, rootPass string,
	args map[string]string) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("StackScriptID", strconv.Itoa(stackScriptId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("StackScriptUDFResponses", stackScriptUDFResponses)
	u.Add("DistributionID", strconv.Itoa(distributionId))
	u.Add("rootPass", rootPass)
	u.Add("Size", strconv.Itoa(size))
	u.Add("Label", label)
	// add optional parameters
	for k, v := range args {
		u.Add(k, v)
	}
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.createfromstackscript", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Delete disk
func (t *LinodeDiskService) Delete(linodeId int, diskId int) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DiskID", strconv.Itoa(diskId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.delete", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Duplicate Disk
func (t *LinodeDiskService) Duplicate(linodeId int, diskId int) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DiskID", strconv.Itoa(diskId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.duplicate", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Imagize a disk
func (t *LinodeDiskService) Imagize(linodeId int, diskId int, description string, label string) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DiskID", strconv.Itoa(diskId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if description != "" {
		u.Add("Description", description)
	}
	if label != "nil" {
		u.Add("Label", label)
	}
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.imagize", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Resize a disk
func (t *LinodeDiskService) Resize(linodeId int, diskId int, size int) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DiskID", strconv.Itoa(diskId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("size", strconv.Itoa(size))
	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.resize", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}

// Update a disk
func (t *LinodeDiskService) Update(linodeId int, diskId int, label string, isReadOnly bool) (*LinodeDiskJobResponse, error) {
	u := &url.Values{}
	u.Add("DiskID", strconv.Itoa(diskId))
	u.Add("LinodeID", strconv.Itoa(linodeId))
	u.Add("Label", label)
	if isReadOnly {
		u.Add("isReadOnly", "1")
	}

	v := LinodeDiskJobResponse{}
	if err := t.client.do("linode.disk.update", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.DiskJob); err != nil {
		return nil, err
	}
	return &v, nil
}
