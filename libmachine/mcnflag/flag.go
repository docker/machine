package mcnflag

import "fmt"

type Flag interface {
	fmt.Stringer
	Default() interface{}
	Description() string
	EnvVarName() string
}

type StringFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  string
}

func (f StringFlag) String() string {
	return f.Name
}

func (f StringFlag) Default() interface{} {
	return f.Value
}

func (f StringFlag) Description() string {
	return f.Usage
}

func (f StringFlag) EnvVarName() string {
	return f.EnvVar
}

type StringSliceFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  []string
}

func (f StringSliceFlag) String() string {
	return f.Name
}

func (f StringSliceFlag) Default() interface{} {
	return f.Value
}

func (f StringSliceFlag) Description() string {
	return f.Usage
}

func (f StringSliceFlag) EnvVarName() string {
	return f.EnvVar
}

type IntFlag struct {
	Name   string
	Usage  string
	EnvVar string
	Value  int
}

func (f IntFlag) String() string {
	return f.Name
}

func (f IntFlag) Default() interface{} {
	return f.Value
}

func (f IntFlag) Description() string {
	return f.Usage
}

func (f IntFlag) EnvVarName() string {
	return f.EnvVar
}

type BoolFlag struct {
	Name   string
	Usage  string
	EnvVar string
}

func (f BoolFlag) String() string {
	return f.Name
}

func (f BoolFlag) Default() interface{} {
	return nil
}

func (f BoolFlag) Description() string {
	return f.Usage
}

func (f BoolFlag) EnvVarName() string {
	return f.EnvVar
}
