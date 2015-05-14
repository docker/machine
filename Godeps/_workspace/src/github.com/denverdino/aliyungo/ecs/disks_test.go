package ecs

import (
	"testing"
)

func TestDisks(t *testing.T) {

	client := NewClient(TestAccessKeyId, TestAccessKeySecret)

	instance, err := client.DescribeInstanceAttribute(TestInstanceId)
	if err != nil {
		t.Fatalf("Failed to DescribeInstanceAttribute for instance %s: %v", TestInstanceId, err)
	}

	args := DescribeDisksArgs{}

	args.InstanceId = TestInstanceId
	args.RegionId = instance.RegionId
	disks, _, err := client.DescribeDisks(&args)

	if err != nil {
		t.Fatalf("Failed to DescribeDisks for instance %s: %v", TestInstanceId, err)
	}

	for _, disk := range disks {
		t.Logf("Disk of instance %s: %++v", TestInstanceId, disk)
	}
}

func TestDiskCreationAndDeletion(t *testing.T) {

	if TestIAmRich == false { //Avoid payment
		return
	}

	client := NewClient(TestAccessKeyId, TestAccessKeySecret)

	instance, err := client.DescribeInstanceAttribute(TestInstanceId)
	if err != nil {
		t.Fatalf("Failed to DescribeInstanceAttribute for instance %s: %v", TestInstanceId, err)
	}

	args := CreateDiskArgs{
		RegionId: instance.RegionId,
		ZoneId:   instance.ZoneId,
		DiskName: "test-disk",
		Size:     5,
	}

	diskId, err := client.CreateDisk(&args)
	if err != nil {
		t.Fatalf("Failed to create disk: %v", err)
	}
	t.Logf("Create disk %s successfully", diskId)

	attachArgs := AttachDiskArgs{
		InstanceId: instance.InstanceId,
		DiskId:     diskId,
	}

	err = client.AttachDisk(&attachArgs)
	if err != nil {
		t.Errorf("Failed to create disk: %v", err)
	} else {
		t.Logf("Attach disk %s to instance %s successfully", diskId, instance.InstanceId)

		instance, err = client.DescribeInstanceAttribute(TestInstanceId)
		if err != nil {
			t.Errorf("Failed to DescribeInstanceAttribute for instance %s: %v", TestInstanceId, err)
		} else {
			t.Logf("Instance: %++v  %v", instance, err)
		}
		err = client.WaitForDisk(instance.RegionId, diskId, DiskStatusInUse, 0)
		if err != nil {
			t.Fatalf("Failed to wait for disk %s to status %s: %v", diskId, DiskStatusInUse, err)
		}
		err = client.DetachDisk(instance.InstanceId, diskId)
		if err != nil {
			t.Errorf("Failed to detach disk: %v", err)
		} else {
			t.Logf("Detach disk %s to instance %s successfully", diskId, instance.InstanceId)
		}

		err = client.WaitForDisk(instance.RegionId, diskId, DiskStatusAvailable, 0)
		if err != nil {
			t.Fatalf("Failed to wait for disk %s to status %s: %v", diskId, DiskStatusAvailable, err)
		}
	}
	err = client.DeleteDisk(diskId)
	if err != nil {
		t.Fatalf("Failed to delete disk %s: %v", diskId, err)
	}
	t.Logf("Delete disk %s successfully", diskId)
}
