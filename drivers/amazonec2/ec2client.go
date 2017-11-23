package amazonec2

import "github.com/aws/aws-sdk-go/service/ec2"

// SecurityGroup defines a set of functions to handle EC2 SecurityGroups
type SecurityGroup interface {
	CreateSecurityGroup(*ec2.CreateSecurityGroupInput) (*ec2.CreateSecurityGroupOutput, error)
	AuthorizeSecurityGroupIngress(*ec2.AuthorizeSecurityGroupIngressInput) (*ec2.AuthorizeSecurityGroupIngressOutput, error)
	DescribeSecurityGroups(*ec2.DescribeSecurityGroupsInput) (*ec2.DescribeSecurityGroupsOutput, error)
	DeleteSecurityGroup(*ec2.DeleteSecurityGroupInput) (*ec2.DeleteSecurityGroupOutput, error)
}

// KeyPair defines a set of functions to handle EC2 KeyPairs
type KeyPair interface {
	DeleteKeyPair(*ec2.DeleteKeyPairInput) (*ec2.DeleteKeyPairOutput, error)
	ImportKeyPair(*ec2.ImportKeyPairInput) (*ec2.ImportKeyPairOutput, error)
	DescribeKeyPairs(*ec2.DescribeKeyPairsInput) (*ec2.DescribeKeyPairsOutput, error)
}

// Instance defines a set of functions to handle EC2 Instaces
type Instance interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
	StartInstances(*ec2.StartInstancesInput) (*ec2.StartInstancesOutput, error)
	RebootInstances(*ec2.RebootInstancesInput) (*ec2.RebootInstancesOutput, error)
	StopInstances(*ec2.StopInstancesInput) (*ec2.StopInstancesOutput, error)
	RunInstances(*ec2.RunInstancesInput) (*ec2.Reservation, error)
	TerminateInstances(*ec2.TerminateInstancesInput) (*ec2.TerminateInstancesOutput, error)
}

// SpotInstances defines a set of functions to handle EC2 SecurityGroups
type SpotInstances interface {
	RequestSpotInstances(*ec2.RequestSpotInstancesInput) (*ec2.RequestSpotInstancesOutput, error)
	DescribeSpotInstanceRequests(*ec2.DescribeSpotInstanceRequestsInput) (*ec2.DescribeSpotInstanceRequestsOutput, error)
	WaitUntilSpotInstanceRequestFulfilled(*ec2.DescribeSpotInstanceRequestsInput) error
}

// Meta defines a set of functions to handle EC2 meta data
type Meta interface {
	DescribeAccountAttributes(*ec2.DescribeAccountAttributesInput) (*ec2.DescribeAccountAttributesOutput, error)
	DescribeSubnets(*ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error)
	CreateTags(*ec2.CreateTagsInput) (*ec2.CreateTagsOutput, error)
}

// Ec2Client defines a set of interfaces to handle EC2 operations
type Ec2Client interface {
	Meta
	//SecurityGroup
	SecurityGroup
	//KeyPair
	KeyPair
	//Instance
	Instance
	//SpotInstances
	SpotInstances
}
