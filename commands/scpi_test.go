package commands

import (
	"testing"

	"github.com/docker/machine/drivers/fakedriver"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/libmachinetest"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
)

func TestGetScpiHostAndImage(t *testing.T) {
	redisImage := "redis:2.8.23"
	redisLatestImage := "redis:latest"
	defautRedisSrc := "default:" + redisImage
	defautRedisLatestSrc := "default:" + redisLatestImage

	defaultHost := &host.Host{
		Name: "default",
		Driver: &fakedriver.Driver{
			MockState: state.Stopped,
		},
	}
	destvmHost := &host.Host{
		Name: "destvm",
		Driver: &fakedriver.Driver{
			MockState: state.Stopped,
		},
	}

	api := &libmachinetest.FakeAPI{
		Hosts: []*host.Host{defaultHost, destvmHost},
	}

	testCases := []struct {
		src      string
		isSource bool
		api      libmachine.API
		host     *host.Host
		image    string
	}{
		{
			src:      defautRedisSrc,
			isSource: true,
			api:      api,
			host:     defaultHost,
			image:    redisImage,
		},
		{
			src:      defautRedisLatestSrc,
			isSource: true,
			api:      api,
			host:     defaultHost,
			image:    redisLatestImage,
		},
	}

	for _, tc := range testCases {
		host, image, err := getScpiHostAndImage(tc.src, tc.isSource, tc.api)
		assert.Equal(t, host, tc.host)
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
