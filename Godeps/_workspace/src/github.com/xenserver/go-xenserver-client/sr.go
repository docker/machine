package client

import (
	"fmt"
	"github.com/nilshell/xmlrpc"
)

type SR XenAPIObject

func (self *SR) GetUuid() (uuid string, err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "SR.get_uuid", self.Ref)
	if err != nil {
		return "", err
	}
	uuid = result.Value.(string)
	return uuid, nil
}

func (self *SR) CreateVdi(name_label string, size int64) (vdi *VDI, err error) {
	vdi = new(VDI)

	vdi_rec := make(xmlrpc.Struct)
	vdi_rec["name_label"] = name_label
	vdi_rec["SR"] = self.Ref
	vdi_rec["virtual_size"] = fmt.Sprintf("%d", size)
	vdi_rec["type"] = "user"
	vdi_rec["sharable"] = false
	vdi_rec["read_only"] = false

	oc := make(xmlrpc.Struct)
	oc["temp"] = "temp"
	vdi_rec["other_config"] = oc

	result := APIResult{}
	err = self.Client.APICall(&result, "VDI.create", vdi_rec)
	if err != nil {
		return nil, err
	}

	vdi.Ref = result.Value.(string)
	vdi.Client = self.Client

	return
}
