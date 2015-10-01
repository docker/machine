package ecs

// Region represents ECS region
type Region string

// Constants of region definition
const (
	Hangzhou = Region("cn-hangzhou")
	Qingdao  = Region("cn-qingdao")
	Beijing  = Region("cn-beijing")
	Hongkong = Region("cn-hongkong")
	Shenzhen = Region("cn-shenzhen")
	USWest1  = Region("us-west-1")
)

var ValidRegions = []Region{Hangzhou, Qingdao, Beijing, Shenzhen, Hongkong, USWest1}

type DescribeRegionsArgs struct {
}

//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/datatype&regiontype
type RegionType struct {
	RegionId  Region
	LocalName string
}

type DescribeRegionsResponse struct {
	CommonResponse
	Regions struct {
		Region []RegionType
	}
}

// DescribeRegions describes regions
//
// You can read doc at http://docs.aliyun.com/#/pub/ecs/open-api/region&describeregions
func (client *Client) DescribeRegions() (regions []RegionType, err error) {
	response := DescribeRegionsResponse{}

	err = client.Invoke("DescribeRegions", &DescribeRegionsArgs{}, &response)

	if err != nil {
		return []RegionType{}, err
	}
	return response.Regions.Region, nil
}
