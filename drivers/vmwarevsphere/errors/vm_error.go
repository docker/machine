/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type VMError struct {
	operation string
	vm        string
	reason    string
}

func NewVMError(operation, vm, reason string) error {
	err := VMError{
		vm:        vm,
		operation: operation,
		reason:    reason,
	}
	return &err
}

func (err *VMError) Error() string {
	return fmt.Sprintf("Unable to %s docker host %s: %s", err.operation, err.vm, err.reason)
}
