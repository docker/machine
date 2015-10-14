package google

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: ""})

	assert.Equal(t, []string{"docker-machine"}, tags)
}

func TestAdditionalTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1"})

	assert.Equal(t, []string{"docker-machine", "tag1"}, tags)
}

func TestAdditionalTags(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1,tag2"})

	assert.Equal(t, []string{"docker-machine", "tag1", "tag2"}, tags)
}
