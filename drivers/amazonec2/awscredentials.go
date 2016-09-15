package amazonec2

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

type awsCredentials interface {
	GetCredentials() *credentials.Credentials
}

type defaultAWSCredentials struct {
	AccessKey    string
	SecretKey    string
	SessionToken string
	File         string
	Profile      string
}

func NewAWSCredentials(id, secret, token, filename, profile string) *defaultAWSCredentials {
	creds := defaultAWSCredentials{
		AccessKey: id,
		SecretKey: secret,
		SessionToken: token,
		File: filename,
		Profile: profile,
	}
	return &creds
}

func (c *defaultAWSCredentials) GetCredentials() *credentials.Credentials {
	providers := []credentials.Provider{}
	if c.AccessKey != "" && c.SecretKey != "" {
		providers = append(providers, &credentials.StaticProvider{Value: credentials.Value{
			AccessKeyID: c.AccessKey,
			SecretAccessKey: c.SecretKey,
			SessionToken: c.SessionToken,
		}})
	}
	if c.File != "" || c.Profile != "" {
		providers = append(providers, &credentials.SharedCredentialsProvider{
			Filename: c.File,
			Profile: c.Profile,
		})
	}
	defaultCreds, err := session.New().Config.Credentials.Get()
	if err == nil {
		providers = append(providers, &credentials.StaticProvider{Value: defaultCreds})
	}
	return credentials.NewChainCredentials(providers)
}
