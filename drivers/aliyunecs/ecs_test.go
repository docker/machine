package aliyunecs

import (
	"testing"
)

func TestRandomPassword(t *testing.T) {

	for i := 0; i < 10; i++ {
		t.Logf("Random Password: %s", randomPassword())
	}

}
