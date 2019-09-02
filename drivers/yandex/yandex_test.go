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

func Test_defaultUserData(t *testing.T) {
	type args struct {
		sshUserName  string
		sshPublicKey string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				sshUserName:  "yc-user",
				sshPublicKey: "public-key-long-string-value",
			},
			want: `#cloud-config
ssh_pwauth: no

users:
  - name: yc-user
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - public-key-long-string-value
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := defaultUserData(tt.args.sshUserName, tt.args.sshPublicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultUserData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("defaultUserData() = %v, want %v", got, tt.want)
			}
		})
	}
}
