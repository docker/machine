package ecs

import (
	"testing"
)

func TestDescribeInstanceTypes(t *testing.T) {

	client := NewClient(TestAccessKeyId, TestAccessKeySecret)
	instanceTypes, err := client.DescribeInstanceTypes()
	if err != nil {
		t.Fatalf("Failed to DescribeInstanceTypes: %v", err)
	}
	for _, instanceType := range instanceTypes {
		t.Logf("InstanceType: %++v", instanceType)
	}
}
