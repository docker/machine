/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

type InvalidLoginError struct {
}

func NewInvalidLoginError() error {
	err := InvalidLoginError{}
	return &err
}

func (err *InvalidLoginError) Error() string {
	return "cannot complete operation due to incorrect vSphere username or password"
}
