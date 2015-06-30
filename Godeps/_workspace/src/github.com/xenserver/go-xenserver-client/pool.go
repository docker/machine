package client

import (
	"github.com/nilshell/xmlrpc"
)

type Pool XenAPIObject

func (self *Pool) GetMaster() (host *Host, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "pool.get_master", self.Ref)
	if err != nil {
		return nil, err
	}
	host = new(Host)
	host.Ref = result.Value.(string)
	host.Client = self.Client
	return host, nil
}

func (self *Pool) GetRecord() (record map[string]interface{}, err error) {
	record = make(map[string]interface{})
	result := APIResult{}
	err = self.Client.APICall(&result, "pool.get_record", self.Ref)
	if err != nil {
		return record, err
	}
	for k, v := range result.Value.(xmlrpc.Struct) {
		record[k] = v
	}
	return record, nil
}
