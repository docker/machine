// API on Network

package ecs

import ()

type AllocatePublicIpAddressArgs struct {
	InstanceId string
}

type AllocatePublicIpAddressResponse struct {
	CommonResponse

	IpAddress string
}

// AllocatePublicIpAddress allocates Public Ip Address
func (client *Client) AllocatePublicIpAddress(instanceId string) (ipAddress string, err error) {
	args := AllocatePublicIpAddressArgs{
		InstanceId: instanceId,
	}
	response := AllocatePublicIpAddressResponse{}
	err = client.Invoke("AllocatePublicIpAddress", &args, &response)
	if err != nil {
		return "", err
	}
	return response.IpAddress, nil
}
