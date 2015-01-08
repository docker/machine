/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type GovcNotFoundError struct {
	path string
}

func NewGovcNotFoundError(path string) error {
	err := GovcNotFoundError{
		path: path,
	}
	return &err
}

func (err *GovcNotFoundError) Error() string {
	return fmt.Sprintf("govc not found: %s", err.path)
}
