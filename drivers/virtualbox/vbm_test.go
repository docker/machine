// +build !race

package virtualbox

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ConcurrentCheckCmder struct {
	calls []func() error
}

func (cc *ConcurrentCheckCmder) Run(stdout, stderr *bytes.Buffer, args ...string) error {
	call := cc.calls[0]
	cc.calls = cc.calls[1:]
	return call()
}

func TestConcurrentVBoxManageSafe(t *testing.T) {
	incr := 0

	calls := []func() error{
		func() error {
			time.Sleep(1 * time.Second)
			incr++
			return nil
		},
		func() error {
			incr++
			return nil
		},
		func() error {
			incr++
			return nil
		},
	}

	v := VBoxCmdManager{
		cmder: &ConcurrentCheckCmder{
			calls: calls,
		},
	}

	for range calls {
		go v.Run(nil, nil)
	}

	time.Sleep(500 * time.Millisecond)

	// Because the first call to v.Run() sleeps for a whole second (which
	// should be plenty of time to hit this block) before updating incr, we
	// should not have updated it at all by now due to the mutex, including
	// in the second call
	assert.Equal(t, 0, incr)
}
