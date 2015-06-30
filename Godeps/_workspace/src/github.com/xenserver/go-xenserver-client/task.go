package client

import (
	"fmt"
	"regexp"
	"strings"
)

type Task XenAPIObject

type TaskStatusType int

const (
	_ TaskStatusType = iota
	Pending
	Success
	Failure
	Cancelling
	Cancelled
)

func (self *Task) GetStatus() (status TaskStatusType, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "task.get_status", self.Ref)
	if err != nil {
		return
	}
	rawStatus := result.Value.(string)
	switch strings.ToLower(rawStatus) {
	case "pending":
		status = Pending
	case "success":
		status = Success
	case "failure":
		status = Failure
	case "cancelling":
		status = Cancelling
	case "cancelled":
		status = Cancelled
	default:
		panic(fmt.Sprintf("Task.get_status: Unknown status '%s'", rawStatus))
	}
	return
}

func (self *Task) GetProgress() (progress float64, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "task.get_progress", self.Ref)
	if err != nil {
		return
	}
	progress = result.Value.(float64)
	return
}

func (self *Task) GetResult() (object *XenAPIObject, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "task.get_result", self.Ref)
	if err != nil {
		return
	}
	switch ref := result.Value.(type) {
	case string:
		// @fixme: xapi currently sends us an xmlrpc-encoded string via xmlrpc.
		// This seems to be a bug in xapi. Remove this workaround when it's fixed
		re := regexp.MustCompile("^<value><array><data><value>([^<]*)</value>.*</data></array></value>$")
		match := re.FindStringSubmatch(ref)
		if match == nil {
			object = nil
		} else {
			object = &XenAPIObject{
				Ref:    match[1],
				Client: self.Client,
			}
		}
	case nil:
		object = nil
	default:
		err = fmt.Errorf("task.get_result: unknown value type %T (expected string or nil)", ref)
	}
	return
}

func (self *Task) GetErrorInfo() (errorInfo []string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "task.get_error_info", self.Ref)
	if err != nil {
		return
	}
	errorInfo = make([]string, 0)
	for _, infoRaw := range result.Value.([]interface{}) {
		errorInfo = append(errorInfo, fmt.Sprintf("%v", infoRaw))
	}
	return
}

func (self *Task) Destroy() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "task.destroy", self.Ref)
	return
}
