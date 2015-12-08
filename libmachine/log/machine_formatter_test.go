package log

import (
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	entry := logrus.NewEntry(logrus.New())
	entry.Message = "foobar"
	formatter := MachineFormatter{}
	bytes, err := formatter.Format(entry)
	assert.Nil(t, err)
	assert.Equal(t, string(bytes[:]), "foobar\n")
}
