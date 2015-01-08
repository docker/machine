/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package errors

import "fmt"

type DatastoreError struct {
	datastore string
	operation string
	reason    string
}

func NewDatastoreError(datastore, operation, reason string) error {
	err := DatastoreError{
		datastore: datastore,
		operation: operation,
		reason:    reason,
	}
	return &err
}

func (err *DatastoreError) Error() string {
	return fmt.Sprintf("Unable to %s on datastore %s due to %s", err.operation, err.datastore, err.reason)
}
