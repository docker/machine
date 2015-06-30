package clcgo

import (
	"encoding/json"
	"errors"
)

const authenticationURL = apiRoot + "/authentication/login"

// APICredentials are used by the Client to identify you to CenturyLink Cloud
// in any request. The BearerToken value will be good for a set period of time,
// defined in the API documentation.
//
// If you anticipate making many API requests in several runs of your
// application, it may be a good idea to save the BearerToken and AccountAlias
// values somewhere so that they can be recalled when they are needed again,
// rather than having to re-authenticate with the API. You can build a
// APICredentials object with those two fields, then assign it to the Client's
// APICredentials value.
//
// The Username and Password values will be present only if you've used
// GetAPICredentials, and can otherwise be ignored.
type APICredentials struct {
	BearerToken  string `json:"bearerToken"`
	AccountAlias string `json:"accountAlias"`
	Username     string `json:"username"` // TODO: nonexistent in get, extract to creation params?
	Password     string `json:"password"` // TODO: nonexistent in get, extract to creation params?
}

// The Client stores your current credentials and uses them to set or fetch
// data from the API. It should be instantiated with the NewClient function.
type Client struct {
	APICredentials APICredentials
	Requestor      requestor
}

func (c APICredentials) RequestForSave(a string) (request, error) {
	return request{URL: authenticationURL, Parameters: c}, nil
}

// NewClient returns a Client that has been configured to communicate to
// CenturyLink Cloud. Before making any requests, the Client must be authorized
// by either setting its APICredentials field or calling the GetAPICredentials
// function.
func NewClient() *Client {
	return &Client{Requestor: clcRequestor{}}
}

// GetAPICredentials accepts username and password strings to populate the
// Client instance with valid APICredentials.
//
// CenturyLink Cloud requires a BearerToken to authorize all requests, and this
// method will fetch one for the user and associate it with the Client. Any
// further requests made by the Client will include that BearerToken.
func (c *Client) GetAPICredentials(u string, p string) error {
	c.APICredentials = APICredentials{Username: u, Password: p}
	_, err := c.SaveEntity(&c.APICredentials)
	if err != nil {
		if rerr, ok := err.(RequestError); ok && rerr.StatusCode == 400 {
			err = errors.New("there was a problem with your credentials")
		}

		return err
	}

	return nil
}

// GetEntity is used to fetch a summary of a resource. When you pass a pointer
// to your resource, GetEntity will set its fields appropriately.
//
// Different resources have different field requirements before their summaries
// can be fetched successfully. If you omit an ID field from a Server, for
// instance, GetEntity will return an error informing you of the missing field.
// An error from GetEntity likely means that your passed Entity was not
// modified.
func (c *Client) GetEntity(e Entity) error {
	url, err := e.URL(c.APICredentials.AccountAlias)
	if err != nil {
		return err
	}
	j, err := c.Requestor.GetJSON(c.APICredentials.BearerToken, request{URL: url})
	if err != nil {
		return err
	}

	return json.Unmarshal(j, &e)
}

// SaveEntity is used to persist a changed resource to CenturyLink Cloud.
//
// Beyond the fields absolutely required to form valid URLs, the presence or
// format of the fields on your resources are not validated before they are
// submitted.  You should check the CenturyLink Cloud API documentation for
// this information. If your submission was unsuccessful, it is likely that the
// error returned is a RequestError, which may contain helpful error messages
// you can use to determine what went wrong.
//
// Calling HasSucceeded on the returned Status will tell you if the resource is
// ready. Some resources are available immediately, but most are not. If the
// resource implements CreationStatusProvidingEntity it will likely take time,
// but regardless you can check with the returned Status to be sure.
//
// The Status does not update itself, and you will need to call GetEntity on it
// periodically to determine when its resource is ready if it was not
// immediately successful.
//
// Never ignore the error result from this call! Even if your code is perfect
// (congratulations, by the way), errors can still occur due to unexpected
// server states or networking issues.
func (c *Client) SaveEntity(e SavableEntity) (Status, error) {
	req, err := e.RequestForSave(c.APICredentials.AccountAlias)
	if err != nil {
		return Status{}, err
	}
	resp, err := c.Requestor.PostJSON(c.APICredentials.BearerToken, req)
	if err != nil {
		return Status{}, err
	}

	if spe, ok := e.(CreationStatusProvidingEntity); ok {
		status, err := spe.StatusFromCreateResponse(resp)
		if err != nil {
			return Status{}, err
		}

		return status, nil
	}

	json.Unmarshal(resp, &e)
	return Status{Status: successfulStatus}, nil
}

// DeleteEntity is used to tell CenturyLink Cloud to remove a resource.
//
// The method returns a Status, which can be used as described in the
// SaveEntity documentation to determine when the work completes. The Entity
// object will not be modified.
func (c *Client) DeleteEntity(e Entity) (Status, error) {
	url, err := e.URL(c.APICredentials.AccountAlias)
	if err != nil {
		return Status{}, err
	}
	resp, err := c.Requestor.DeleteJSON(c.APICredentials.BearerToken, request{URL: url})
	if err != nil {
		return Status{}, err
	}

	if spe, ok := e.(DeletionStatusProvidingEntity); ok {
		status, err := spe.StatusFromDeleteResponse(resp)
		if err != nil {
			return Status{}, err
		}

		return status, nil
	}

	return Status{Status: successfulStatus}, nil
}
