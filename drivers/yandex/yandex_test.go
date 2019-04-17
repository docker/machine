package yandex

import (
	"reflect"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"yandex-folder-id": "FOLDER_ID",
			"yandex-token":     "TOKEN",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestDriver_ParsedLabels(t *testing.T) {
	tests := []struct {
		name   string
		labels []string
		want   map[string]string
	}{
		{
			name:   "one label",
			labels: []string{"somekey=somevalue"},
			want: map[string]string{
				"somekey": "somevalue",
			},
		},
		{
			name:   "several labels",
			labels: []string{"somekey=somevalue", "foo=bar"},
			want: map[string]string{
				"somekey": "somevalue",
				"foo":     "bar",
			},
		},
		{
			name:   "partial labels",
			labels: []string{"somekey=", "foo=bar", "wo-value="},
			want: map[string]string{
				"somekey":  "",
				"foo":      "bar",
				"wo-value": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Driver{
				Labels: tt.labels,
			}
			d.ParsedLabels()
			if got := d.ParsedLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Driver.ParsedLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}
