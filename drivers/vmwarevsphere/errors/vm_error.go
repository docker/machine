/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type VmError struct {
	operation string
	vm        string
	reason    string
}

func NewVmError(operation, vm, reason string) error {
	err := VmError{
		vm:        vm,
		operation: operation,
		reason:    reason,
	}
	return &err
}

func (err *VmError) Error() string {
	return fmt.Sprintf("Unable to %s docker host %s: %s", err.operation, err.vm, err.reason)
}
