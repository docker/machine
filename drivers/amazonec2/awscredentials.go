package amazonec2

import "github.com/aws/aws-sdk-go/aws/credentials"
import "github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
import "github.com/aws/aws-sdk-go/aws/session"

type awsCredentials interface {
	NewStaticCredentials(id, secret, token string) *credentials.Credentials

	NewSharedCredentials(filename, profile string) *credentials.Credentials

	NewRoleCredentials() *credentials.Credentials
}

type defaultAWSCredentials struct{}

func (c *defaultAWSCredentials) NewStaticCredentials(id, secret, token string) *credentials.Credentials {
	return credentials.NewStaticCredentials(id, secret, token)
}

func (c *defaultAWSCredentials) NewSharedCredentials(filename, profile string) *credentials.Credentials {
	return credentials.NewSharedCredentials(filename, profile)
}

func (c *defaultAWSCredentials) NewRoleCredentials() *credentials.Credentials {
	return ec2rolecreds.NewCredentials(session.New())
}
