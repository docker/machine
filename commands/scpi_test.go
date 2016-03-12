package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseScpiTargetInfo(t *testing.T) {
	redisImage := "redis:2.8.23"
	redisLatestImage := "redis:latest"
	defautRedisSrc := "default:" + redisImage
	defautRedisLatestSrc := "default:" + redisLatestImage
	somevm := "somevm"

	hostInfoLoader := &MockHostInfoLoader{MockHostInfo{
		sshKeyPath: "/fake/keypath/id_rsa",
	}}

	testCases := []struct {
		srcOrDest      string
		isSource       bool
		hostInfoLoader HostInfoLoader
		hostName       string
		image          string
	}{
		{
			srcOrDest:      defautRedisSrc,
			isSource:       true,
			hostInfoLoader: hostInfoLoader,
			hostName:       "default",
			image:          redisImage,
		},
		{
			srcOrDest:      defautRedisLatestSrc,
			isSource:       true,
			hostInfoLoader: hostInfoLoader,
			hostName:       "default",
			image:          redisLatestImage,
		},
		{
			srcOrDest:      somevm,
			isSource:       false,
			hostInfoLoader: hostInfoLoader,
			hostName:       "somevm",
			image:          "",
		},
	}

	for _, tc := range testCases {
		hostName, image, err := parseScpiTargetInfo(tc.srcOrDest, tc.isSource, tc.hostInfoLoader)
		assert.Equal(t, hostName, tc.hostName)
		assert.Equal(t, image, tc.image)
		assert.Equal(t, err, nil)
	}
}

func TestGenerateSaveFilename(t *testing.T) {
	assert.Equal(t, generateSaveFilename("redis:2.8.23"), "/tmp/redis__2.8.23.tar")
	assert.Equal(t, generateSaveFilename("redis:latest"), "/tmp/redis__latest.tar")
	assert.Equal(t, generateSaveFilename("nsqio/nsq:v0.3.6"), "/tmp/nsqio_nsq__v0.3.6.tar")
	assert.Equal(t, generateSaveFilename("nsqio/nsq:latest"), "/tmp/nsqio_nsq__latest.tar")
}
