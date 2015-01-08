/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type InvalidStateError struct {
	vm string
}

func NewInvalidStateError(vm string) error {
	err := InvalidStateError{
		vm: vm,
	}
	return &err
}

func (err *InvalidStateError) Error() string {
	return fmt.Sprintf("Machine %s state invalid", err.vm)
}
