package log

import (
	"bytes"

	"github.com/Sirupsen/logrus"
)

type MachineFormatter struct {
}

func (d *MachineFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}

	b.WriteString(entry.Message)
	b.WriteByte('\n')

	return b.Bytes(), nil
}
