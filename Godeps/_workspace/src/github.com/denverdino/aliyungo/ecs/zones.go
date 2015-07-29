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

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&availableresourcecreationtype
type AvailableResourceCreationType struct {
	ResourceTypes []ResourceType //enum for Instance, Disk, VSwitch
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&availablediskcategoriestype
type AvailableDiskCategoriesType struct {
	DiskCategories []DiskCategory //enum for cloud, ephemeral, ephemeral_ssd
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&zonetype
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
