package ecs

import (
	"testing"
)

func TestAllocatePublicIpAddress(t *testing.T) {

	client := NewClient(TestAccessKeyId, TestAccessKeySecret)
	instance, err := client.DescribeInstanceAttribute(TestInstanceId)
	if err != nil {
		t.Fatalf("Failed to describe instance %s: %v", TestInstanceId, err)
	}
	t.Logf("Instance: %++v  %v", instance, err)
	ipAddr, err := client.AllocatePublicIpAddress(TestInstanceId)
	if err != nil {
		t.Fatalf("Failed to allocate public IP address for instance %s: %v", TestInstanceId, err)
	}
	t.Logf("Public IP address of instance %s: %s", TestInstanceId, ipAddr)

}
