package persist

import (
	"testing"

	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/stretchr/testify/assert"
)

func TestSaveError(t *testing.T) {
	saveError := NewSaveError("machine", errors.New("Reason"))

	assert.Equal(t, `Error attempting to save host "machine" to store: Reason`, saveError.Error())
}
