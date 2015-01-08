/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type IncompleteVsphereConfigError struct {
	component string
}

func NewIncompleteVsphereConfigError(component string) error {
	err := IncompleteVsphereConfigError{
		component: component,
	}
	return &err
}

func (err *IncompleteVsphereConfigError) Error() string {
	return fmt.Sprintf("Incomplete vSphere information: missing %s", err.component)
}
