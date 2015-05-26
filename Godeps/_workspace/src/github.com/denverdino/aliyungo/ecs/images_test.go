package ecs

import (
	"testing"
)

func TestImageCreationAndDeletion(t *testing.T) {

	client := NewClient(TestAccessKeyId, TestAccessKeySecret)

	instance, err := client.DescribeInstanceAttribute(TestInstanceId)
	if err != nil {
		t.Errorf("Failed to DescribeInstanceAttribute for instance %s: %v", TestInstanceId, err)
	}

	args := DescribeSnapshotsArgs{}

	args.InstanceId = TestInstanceId
	args.RegionId = instance.RegionId
	snapshots, _, err := client.DescribeSnapshots(&args)

	if err != nil {
		t.Errorf("Failed to DescribeSnapshots for instance %s: %v", TestInstanceId, err)
	}

	if len(snapshots) > 0 {

		createImageArgs := CreateImageArgs{
			RegionId:   instance.RegionId,
			SnapshotId: snapshots[0].SnapshotId,

			ImageName:    "My_Test_Image_for_AliyunGo",
			ImageVersion: "1.0",
			Description:  "My Test Image for AliyunGo description",
			ClientToken:  client.GenerateClientToken(),
		}
		imageId, err := client.CreateImage(&createImageArgs)
		if err != nil {
			t.Errorf("Failed to CreateImage for SnapshotId %s: %v", createImageArgs.SnapshotId, err)
		}
		t.Logf("Image %s is created successfully.", imageId)

		err = client.DeleteImage(instance.RegionId, imageId)
		if err != nil {
			t.Errorf("Failed to DeleteImage for %s: %v", imageId, err)
		}
		t.Logf("Image %s is deleted successfully.", imageId)

	}
}
