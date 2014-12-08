package gopherstack

import (
	"net/url"
)

// Create a SSH key pair
func (c CloudstackClient) CreateSSHKeyPair(name string) (CreateSshKeyPairResponse, error) {
	var resp CreateSshKeyPairResponse
	params := url.Values{}
	params.Set("name", name)
	response, err := NewRequest(c, "createSSHKeyPair", params)
	if err != nil {
		return resp, err
	}
	resp = response.(CreateSshKeyPairResponse)
	return resp, nil
}

// Deletes an SSH key pair
func (c CloudstackClient) DeleteSSHKeyPair(name string) (DeleteSshKeyPairResponse, error) {
	var resp DeleteSshKeyPairResponse
	params := url.Values{}
	params.Set("name", name)
	response, err := NewRequest(c, "deleteSSHKeyPair", params)
	if err != nil {
		return resp, err
	}
	resp = response.(DeleteSshKeyPairResponse)
	return resp, err
}

type CreateSshKeyPairResponse struct {
	Createsshkeypairresponse struct {
		Keypair struct {
			Fingerprint string `json:"fingerprint"`
			Name        string `json:"name"`
			Privatekey  string `json:"privatekey"`
		} `json:"keypair"`
	} `json:"createsshkeypairresponse"`
}

type DeleteSshKeyPairResponse struct {
	Deletesshkeypairresponse struct {
		Success string `json:"success"`
	} `json:"deletesshkeypairresponse"`
}
