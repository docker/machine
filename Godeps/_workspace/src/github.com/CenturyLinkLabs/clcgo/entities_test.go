package clcgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusURL(t *testing.T) {
	s := Status{URI: "/v2/status/1234"}
	url, err := s.URL("AA")
	assert.NoError(t, err)
	assert.Equal(t, apiDomain+"/v2/status/1234", url)
}

func TestStatusHasSucceeded(t *testing.T) {
	s := Status{}
	assert.False(t, s.HasSucceeded())

	s.Status = "executing"
	assert.False(t, s.HasSucceeded())

	s.Status = "succeeded"
	assert.True(t, s.HasSucceeded())
}
