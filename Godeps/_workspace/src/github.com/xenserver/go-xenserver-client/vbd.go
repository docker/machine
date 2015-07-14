package client

import (
	"github.com/nilshell/xmlrpc"
)

type VBD XenAPIObject

func (self *VBD) GetRecord() (record map[string]interface{}, err error) {
	record = make(map[string]interface{})
	result := APIResult{}
	err = self.Client.APICall(&result, "VBD.get_record", self.Ref)
	if err != nil {
		return record, err
	}
	for k, v := range result.Value.(xmlrpc.Struct) {
		record[k] = v
	}
	return record, nil
}

func (self *VBD) GetVDI() (vdi *VDI, err error) {
	vbd_rec, err := self.GetRecord()
	if err != nil {
		return nil, err
	}

	vdi = new(VDI)
	vdi.Ref = vbd_rec["VDI"].(string)
	vdi.Client = self.Client

	return vdi, nil
}

func (self *VBD) Eject() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VBD.eject", self.Ref)
	if err != nil {
		return err
	}
	return nil
}

func (self *VBD) Unplug() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VBD.unplug", self.Ref)
	if err != nil {
		return err
	}
	return nil
}

func (self *VBD) Destroy() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VBD.destroy", self.Ref)
	if err != nil {
		return err
	}
	return nil
}
