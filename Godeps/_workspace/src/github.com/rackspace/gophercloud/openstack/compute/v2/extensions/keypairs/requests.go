package keypairs

import (
	"errors"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/pagination"
)

// List returns a Pager that allows you to iterate over a collection of KeyPairs.
func List(client *gophercloud.ServiceClient) pagination.Pager {
	return pagination.NewPager(client, listURL(client), func(r pagination.PageResult) pagination.Page {
		return KeyPairPage{pagination.SinglePageBase(r)}
	})
}

// CreateOptsBuilder describes struct types that can be accepted by the Create call. Notable, the
// CreateOpts struct in this package does.
type CreateOptsBuilder interface {
	ToKeyPairCreateMap() (map[string]interface{}, error)
}

// CreateOpts species keypair creation or import parameters.
type CreateOpts struct {
	// Name [required] is a friendly name to refer to this KeyPair in other services.
	Name string

	// PublicKey [optional] is a pregenerated OpenSSH-formatted public key. If provided, this key
	// will be imported and no new key will be created.
	PublicKey string
}

// ToKeyPairCreateMap constructs a request body from CreateOpts.
func (opts CreateOpts) ToKeyPairCreateMap() (map[string]interface{}, error) {
	if opts.Name == "" {
		return nil, errors.New("Missing field required for keypair creation: Name")
	}

	keypair := make(map[string]interface{})
	keypair["name"] = opts.Name
	if opts.PublicKey != "" {
		keypair["public_key"] = opts.PublicKey
	}

	return map[string]interface{}{"keypair": keypair}, nil
}

// Create requests the creation of a new keypair on the server, or to import a pre-existing
// keypair.
func Create(client *gophercloud.ServiceClient, opts CreateOptsBuilder) CreateResult {
	var res CreateResult

	reqBody, err := opts.ToKeyPairCreateMap()
	if err != nil {
		res.Err = err
		return res
	}

	_, res.Err = perigee.Request("POST", createURL(client), perigee.Options{
		MoreHeaders: client.AuthenticatedHeaders(),
		ReqBody:     reqBody,
		Results:     &res.Body,
		OkCodes:     []int{200},
	})
	return res
}

// Get returns public data about a previously uploaded KeyPair.
func Get(client *gophercloud.ServiceClient, name string) GetResult {
	var res GetResult
	_, res.Err = perigee.Request("GET", getURL(client, name), perigee.Options{
		MoreHeaders: client.AuthenticatedHeaders(),
		Results:     &res.Body,
		OkCodes:     []int{200},
	})
	return res
}

// Delete requests the deletion of a previous stored KeyPair from the server.
func Delete(client *gophercloud.ServiceClient, name string) DeleteResult {
	var res DeleteResult
	_, res.Err = perigee.Request("DELETE", deleteURL(client, name), perigee.Options{
		MoreHeaders: client.AuthenticatedHeaders(),
		OkCodes:     []int{202},
	})
	return res
}
