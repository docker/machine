package client

type VIF XenAPIObject

func (self *VIF) Destroy() (err error) {
	result := APIResult{}
	err = self.Client.APICall(&result, "VIF.destroy", self.Ref)
	if err != nil {
		return err
	}
	return nil
}

func (self *VIF) GetNetwork() (network *Network, err error) {

	network = new(Network)
	result := APIResult{}
	err = self.Client.APICall(&result, "VIF.get_network", self.Ref)

	if err != nil {
		return nil, err
	}
	network.Ref = result.Value.(string)
	network.Client = self.Client
	return

}
