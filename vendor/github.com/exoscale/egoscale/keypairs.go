package egoscale

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
)

// SSHKeyPair represents an SSH key pair
type SSHKeyPair struct {
	Account     string `json:"account,omitempty"` // must be used with a Domain ID
	DomainID    string `json:"domainid,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Name        string `json:"name,omitempty"`
	PrivateKey  string `json:"privatekey,omitempty"`
}

// Get populates the given SSHKeyPair
func (ssh *SSHKeyPair) Get(ctx context.Context, client *Client) error {
	resp, err := client.RequestWithContext(ctx, &ListSSHKeyPairs{
		Account:     ssh.Account,
		DomainID:    ssh.DomainID,
		Name:        ssh.Name,
		Fingerprint: ssh.Fingerprint,
		ProjectID:   ssh.ProjectID,
	})

	if err != nil {
		return err
	}

	sshs := resp.(*ListSSHKeyPairsResponse)
	count := len(sshs.SSHKeyPair)
	if count == 0 {
		return &ErrorResponse{
			ErrorCode: ParamError,
			ErrorText: fmt.Sprintf("SSHKeyPair not found"),
		}
	} else if count > 1 {
		return fmt.Errorf("More than one SSHKeyPair was found")
	}

	return copier.Copy(ssh, sshs.SSHKeyPair[0])
}

// Delete removes the given SSH key, by Name
func (ssh *SSHKeyPair) Delete(ctx context.Context, client *Client) error {
	if ssh.Name == "" {
		return fmt.Errorf("An SSH Key Pair may only be deleted using Name")
	}

	return client.BooleanRequestWithContext(ctx, &DeleteSSHKeyPair{
		Name:      ssh.Name,
		Account:   ssh.Account,
		DomainID:  ssh.DomainID,
		ProjectID: ssh.ProjectID,
	})
}

// CreateSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/createSSHKeyPair.html
type CreateSSHKeyPair struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*CreateSSHKeyPair) APIName() string {
	return "createSSHKeyPair"
}

func (*CreateSSHKeyPair) response() interface{} {
	return new(CreateSSHKeyPairResponse)
}

// CreateSSHKeyPairResponse represents the creation of an SSH Key Pair
type CreateSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// DeleteSSHKeyPair represents a new keypair to be created
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/deleteSSHKeyPair.html
type DeleteSSHKeyPair struct {
	Name      string `json:"name"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*DeleteSSHKeyPair) APIName() string {
	return "deleteSSHKeyPair"
}

func (*DeleteSSHKeyPair) response() interface{} {
	return new(booleanSyncResponse)
}

// RegisterSSHKeyPair represents a new registration of a public key in a keypair
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/registerSSHKeyPair.html
type RegisterSSHKeyPair struct {
	Name      string `json:"name"`
	PublicKey string `json:"publickey"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*RegisterSSHKeyPair) APIName() string {
	return "registerSSHKeyPair"
}

func (*RegisterSSHKeyPair) response() interface{} {
	return new(RegisterSSHKeyPairResponse)
}

// RegisterSSHKeyPairResponse represents the creation of an SSH Key Pair
type RegisterSSHKeyPairResponse struct {
	KeyPair SSHKeyPair `json:"keypair"`
}

// ListSSHKeyPairs represents a query for a list of SSH KeyPairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/listSSHKeyPairs.html
type ListSSHKeyPairs struct {
	Account     string `json:"account,omitempty"`
	DomainID    string `json:"domainid,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	IsRecursive *bool  `json:"isrecursive,omitempty"`
	Keyword     string `json:"keyword,omitempty"`
	ListAll     *bool  `json:"listall,omitempty"`
	Name        string `json:"name,omitempty"`
	Page        int    `json:"page,omitempty"`
	PageSize    int    `json:"pagesize,omitempty"`
	ProjectID   string `json:"projectid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*ListSSHKeyPairs) APIName() string {
	return "listSSHKeyPairs"
}

func (*ListSSHKeyPairs) response() interface{} {
	return new(ListSSHKeyPairsResponse)
}

// ListSSHKeyPairsResponse represents a list of SSH key pairs
type ListSSHKeyPairsResponse struct {
	Count      int          `json:"count"`
	SSHKeyPair []SSHKeyPair `json:"sshkeypair"`
}

// ResetSSHKeyForVirtualMachine (Async) represents a change for the key pairs
//
// CloudStack API: http://cloudstack.apache.org/api/apidocs-4.10/apis/resetSSHKeyForVirtualMachine.html
type ResetSSHKeyForVirtualMachine struct {
	ID        string `json:"id"`
	KeyPair   string `json:"keypair"`
	Account   string `json:"account,omitempty"`
	DomainID  string `json:"domainid,omitempty"`
	ProjectID string `json:"projectid,omitempty"`
}

// APIName returns the CloudStack API command name
func (*ResetSSHKeyForVirtualMachine) APIName() string {
	return "resetSSHKeyForVirtualMachine"
}

func (*ResetSSHKeyForVirtualMachine) asyncResponse() interface{} {
	return new(ResetSSHKeyForVirtualMachineResponse)
}

// ResetSSHKeyForVirtualMachineResponse represents the modified VirtualMachine
type ResetSSHKeyForVirtualMachineResponse VirtualMachineResponse
