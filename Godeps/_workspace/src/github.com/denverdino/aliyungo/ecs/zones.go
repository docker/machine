package ecs

type ResourceType string

const (
	ResourceTypeInstance = ResourceType("Instance")
	ResourceTypeDisk     = ResourceType("Disk")
	ResourceTypeVSwitch  = ResourceType("VSwitch")
)

type DescribeZonesArgs struct {
	RegionId Region
}

type AvailableResourceCreationType struct {
	ResourceTypes []ResourceType //enum for Instance, Disk, VSwitch
}

type AvailableDiskCategoriesType struct {
	DiskCategories []DiskCategory //enum for cloud, ephemeral, ephemeral_ssd
}

type ZoneType struct {
	ZoneId                    string
	LocalName                 string
	AvailableResourceCreation AvailableResourceCreationType
	AvailableDiskCategories   AvailableDiskCategoriesType
}

type DescribeZonesResponse struct {
	CommonResponse
	Zones struct {
		Zone []ZoneType
	}
}

// DescribeZones describes zones
func (client *Client) DescribeZones(regionId Region) (zones []ZoneType, err error) {
	args := DescribeZonesArgs{
		RegionId: regionId,
	}
	response := DescribeZonesResponse{}

	err = client.Invoke("DescribeZones", &args, &response)

	if err == nil {
		return response.Zones.Zone, nil
	}

	return []ZoneType{}, err
}
