package google

import (
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("", "")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"google-project": "PROJECT",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestMetadataMapFromStringSlice(t *testing.T) {
	tests := map[string]struct {
		slice          []string
		expectedResult metadataMap
	}{
		"empty slice": {
			slice:          []string{""},
			expectedResult: metadataMap{},
		},
		"missing key=value pair": {
			slice:          []string{"key_1"},
			expectedResult: metadataMap{},
		},
		"key=value pair present": {
			slice:          []string{"key_1=value_1"},
			expectedResult: metadataMap{"key_1": "value_1"},
		},
		"multiple = characters present": {
			slice:          []string{"key_1=value=_1="},
			expectedResult: metadataMap{"key_1": "value=_1="},
		},
		"invalid key value pair": {
			slice:          []string{"key_1="},
			expectedResult: metadataMap{"key_1": ""},
		},
		"multiple metadata items": {
			slice: []string{"key_1=value_1", "key_2=value_2"},
			expectedResult: metadataMap{
				"key_1": "value_1",
				"key_2": "value_2",
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			result := metadataMapFromStringSlice(tt.slice)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
