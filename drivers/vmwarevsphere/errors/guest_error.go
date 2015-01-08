/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type GuestError struct {
	vm        string
	operation string
	reason    string
}

func NewGuestError(vm, operation, reason string) error {
	err := GuestError{
		vm:        vm,
		operation: operation,
		reason:    reason,
	}
	return &err
}

func (err *GuestError) Error() string {
	return fmt.Sprintf("Unable to %s on vm %s due to %s", err.operation, err.vm, err.reason)
}
