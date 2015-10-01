package ecs

import (
	"testing"
)

func TestGenerateClientToken(t *testing.T) {
	client := NewClient(TestAccessKeyId, TestAccessKeySecret)
	for i := 0; i < 10; i++ {
		t.Log("GenerateClientToken: ", client.GenerateClientToken())
	}

}

func TestECSDescribe(t *testing.T) {
	client := NewClient(TestAccessKeyId, TestAccessKeySecret)

	regions, err := client.DescribeRegions()

	t.Log("regions: ", regions, err)

	for _, region := range regions {
		zones, err := client.DescribeZones(region.RegionId)
		t.Log("zones: ", zones, err)
		for _, zone := range zones {
			args := DescribeInstanceStatusArgs{
				RegionId: region.RegionId,
				ZoneId:   zone.ZoneId,
			}
			instanceStatuses, pagination, err := client.DescribeInstanceStatus(&args)
			t.Log("instanceStatuses: ", instanceStatuses, pagination, err)
			for _, instanceStatus := range instanceStatuses {
				instance, err := client.DescribeInstanceAttribute(instanceStatus.InstanceId)
				t.Logf("Instance: %++v", instance)
				t.Logf("Error: %++v", err)
			}
			args1 := DescribeInstancesArgs{
				RegionId: region.RegionId,
				ZoneId:   zone.ZoneId,
			}
			instances, _, err := client.DescribeInstances(&args1)
			if err != nil {
				t.Errorf("Failed to describe instance %s %s", region.RegionId, zone.ZoneId)
			} else {
				for _, instance := range instances {
					t.Logf("Instance: %++v", instance)
				}
			}

		}
		args := DescribeImagesArgs{RegionId: region.RegionId}

		for {

			images, pagination, err := client.DescribeImages(&args)
			if err != nil {
				t.Fatalf("Failed to describe images: %v", err)
				break
			} else {
				t.Logf("Total image count for region %s: %d", region.RegionId, pagination.TotalCount)
				for _, image := range images {
					t.Logf("Image: %++v", image)
				}
				nextPage := pagination.NextPage()
				if nextPage == nil {
					break
				}
				args.Pagination = *nextPage
			}
		}
	}
}
