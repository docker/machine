package amazonec2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAccessKeyIsMandatoryWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "", "", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestAccessKeyIsMandatoryEvenIfSecretKeyIsPassedWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("", "secret", "", "", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestSecretKeyIsMandatoryWhenSystemCredentialsAreNotPresent(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "", "", "", "")
	awsCreds.fallbackProvider = nil

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}

func TestFallbackCredentialsAreLoadedWhenAccessKeyAndSecretKeyAreMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "", "", "")
	awsCreds.fallbackProvider = fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedWhenAccessKeyIsMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("", "secret", "", "", "")
	awsCreds.fallbackProvider = fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedWhenSecretKeyIsMissing(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "", "", "", "")
	awsCreds.fallbackProvider = fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestOptionCredentialsAreLoadedWhenAccessKeyAndSecretKeyAreProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "", "", "")
	awsCreds.fallbackProvider = fallbackCredentials{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestCredentialsAreLoadedFromFileWhenFilenameIsProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "", "my/aws/credentials", "")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = NewTestFileCredentialsProvider("access", "secret", "")

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestCredentialsAreLoadedFromFileWhenProfileIsProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "", "", "non-default-profile")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = NewTestFileCredentialsProvider("access", "secret", "")

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestCredentialsAreLoadedFromFileWhenFilenameAndProfileAreProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("", "", "", "my/aws/credentials", "non-default-profile")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = NewTestFileCredentialsProvider("access", "secret", "")

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestOptionCredentialsAreLoadedEvenIfFilenameAndProfileAreProvided(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "token", "my/aws/credentials", "non-default-profile")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = NewTestFileCredentialsProvider("foo", "bar", "")

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "token", creds.SessionToken)
}

func TestCredentialsAreLoadedFromFileIfStaticCredentialsGenerateError(t *testing.T) {
	awsCreds := NewAWSCredentials("foo", "bar", "", "my/aws/credentials", "non-default-profile")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = NewTestFileCredentialsProviderWithStaticError("access", "secret", "")

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "access", creds.AccessKeyID)
	assert.Equal(t, "secret", creds.SecretAccessKey)
	assert.Equal(t, "", creds.SessionToken)
}

func TestFallbackCredentialsAreLoadedIfBothStaticAndSharedCredentialsGenerateError(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "token", "my/aws/credentials", "non-default-profile")
	awsCreds.fallbackProvider = fallbackCredentials{}
	awsCreds.providerFactory = errorCredentialsProvider{}

	creds, err := awsCreds.Credentials().Get()

	assert.NoError(t, err)
	assert.Equal(t, "fallback_access", creds.AccessKeyID)
	assert.Equal(t, "fallback_secret", creds.SecretAccessKey)
	assert.Equal(t, "fallback_token", creds.SessionToken)
}

func TestErrorGeneratedWhenAllProvidersGenerateErrors(t *testing.T) {
	awsCreds := NewAWSCredentials("access", "secret", "token", "my/aws/credentials", "non-default-profile")
	awsCreds.fallbackProvider = errorFallbackCredentials{}
	awsCreds.providerFactory = errorCredentialsProvider{}

	_, err := awsCreds.Credentials().Get()
	assert.Error(t, err)
}
