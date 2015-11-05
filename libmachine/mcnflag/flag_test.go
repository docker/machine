package mcnflag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringFlag(t *testing.T) {
	var flag Flag = &StringFlag{
		Name:   "name",
		Usage:  "usage",
		EnvVar: "ENV",
		Value:  "default",
	}

	assert.Equal(t, "name", flag.String())
	assert.Equal(t, "usage", flag.Description())
	assert.Equal(t, "ENV", flag.EnvVarName())
	assert.Equal(t, "default", flag.Default())
}

func TestStringSliceFlag(t *testing.T) {
	var flag Flag = &StringSliceFlag{
		Name:   "name",
		Usage:  "usage",
		EnvVar: "ENV",
		Value:  []string{"value1", "value2"},
	}

	assert.Equal(t, "name", flag.String())
	assert.Equal(t, "usage", flag.Description())
	assert.Equal(t, "ENV", flag.EnvVarName())
	assert.Equal(t, []string{"value1", "value2"}, flag.Default())
}

func TestIntFlag(t *testing.T) {
	var flag Flag = &IntFlag{
		Name:   "name",
		Usage:  "usage",
		EnvVar: "ENV",
		Value:  42,
	}

	assert.Equal(t, "name", flag.String())
	assert.Equal(t, "usage", flag.Description())
	assert.Equal(t, "ENV", flag.EnvVarName())
	assert.Equal(t, 42, flag.Default())
}

func TestBoolFlag(t *testing.T) {
	var flag Flag = &BoolFlag{
		Name:   "name",
		Usage:  "usage",
		EnvVar: "ENV",
	}

	assert.Equal(t, "name", flag.String())
	assert.Equal(t, "usage", flag.Description())
	assert.Equal(t, "ENV", flag.EnvVarName())
	assert.Equal(t, nil, flag.Default())
}
