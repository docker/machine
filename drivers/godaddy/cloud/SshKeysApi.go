package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dghubble/sling"
)

type SshKeysApi struct {
	basePath string
	sling    *sling.Sling
}

func NewSshKeysApi(basePath string, apiKey string) *SshKeysApi {
	s := sling.New().Set("Authorization", fmt.Sprintf("sso-key %s", apiKey))
	return &SshKeysApi{
		basePath: basePath,
		sling:    s,
	}
}

/**
 * Create a new SSH key
 *
 * @param body SSH key details
 * @return SSHKey
 */
//func (a SshKeysApi) AddSSHKey (body SSHKeyCreate) (SSHKey, error) {
func (a SshKeysApi) AddSSHKey(body SSHKeyCreate) (SSHKey, error) {

	_sling := a.sling.New().Post(a.basePath)

	// create path and map variables
	path := "/v1/cloud/sshKeys"

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	// body params
	_sling = _sling.BodyJSON(body)

	var successPayload = new(SSHKey)

	// We use this map (below) so that any arbitrary error JSON can be handled.
	// FIXME: This is in the absence of this Go generator honoring the non-2xx
	// response (error) models, which needs to be implemented at some point.
	var failurePayload map[string]interface{}

	httpResponse, err := _sling.Receive(successPayload, &failurePayload)

	if err == nil {
		// err == nil only means that there wasn't a sub-application-layer error (e.g. no network error)
		if failurePayload != nil {
			// If the failurePayload is present, there likely was some kind of non-2xx status
			// returned (and a JSON payload error present)
			var str []byte
			str, err = json.Marshal(failurePayload)
			if err == nil { // For safety, check for an error marshalling... probably superfluous
				// This will return the JSON error body as a string
				err = errors.New(string(str))
			}
		} else {
			// So, there was no network-type error, and nothing in the failure payload,
			// but we should still check the status code
			if httpResponse == nil {
				// This should never happen...
				err = errors.New("No HTTP Response received.")
			} else if code := httpResponse.StatusCode; 200 > code || code > 299 {
				err = errors.New("HTTP Error: " + string(httpResponse.StatusCode))
			}
		}
	}

	return *successPayload, err
}

/**
 * Delete a SSH key resource
 * Permanently deletes the SSH key, making it unavailable for new servers.
 * @param sshKeyId Id of SSH key to be deleted
 * @return void
 */
//func (a SshKeysApi) DeleteSSHKey (sshKeyId string) (error) {
func (a SshKeysApi) DeleteSSHKey(sshKeyId string) error {

	_sling := a.sling.New().Delete(a.basePath)

	// create path and map variables
	path := "/v1/cloud/sshKeys/{sshKeyId}"
	path = strings.Replace(path, "{"+"sshKeyId"+"}", fmt.Sprintf("%v", sshKeyId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	// We use this map (below) so that any arbitrary error JSON can be handled.
	// FIXME: This is in the absence of this Go generator honoring the non-2xx
	// response (error) models, which needs to be implemented at some point.
	var failurePayload map[string]interface{}

	httpResponse, err := _sling.Receive(nil, &failurePayload)

	if err == nil {
		// err == nil only means that there wasn't a sub-application-layer error (e.g. no network error)
		if failurePayload != nil {
			// If the failurePayload is present, there likely was some kind of non-2xx status
			// returned (and a JSON payload error present)
			var str []byte
			str, err = json.Marshal(failurePayload)
			if err == nil { // For safety, check for an error marshalling... probably superfluous
				// This will return the JSON error body as a string
				err = errors.New(string(str))
			}
		} else {
			// So, there was no network-type error, and nothing in the failure payload,
			// but we should still check the status code
			if httpResponse == nil {
				// This should never happen...
				err = errors.New("No HTTP Response received.")
			} else if code := httpResponse.StatusCode; 200 > code || code > 299 {
				err = errors.New("HTTP Error: " + string(httpResponse.StatusCode))
			}
		}
	}

	return err
}

/**
 * Find SSH key by sshKeyId
 *
 * @param sshKeyId Id of SSH key to be fetched
 * @return SSHKey
 */
//func (a SshKeysApi) GetSSHKeyById (sshKeyId string) (SSHKey, error) {
func (a SshKeysApi) GetSSHKeyById(sshKeyId string) (SSHKey, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/sshKeys/{sshKeyId}"
	path = strings.Replace(path, "{"+"sshKeyId"+"}", fmt.Sprintf("%v", sshKeyId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(SSHKey)

	// We use this map (below) so that any arbitrary error JSON can be handled.
	// FIXME: This is in the absence of this Go generator honoring the non-2xx
	// response (error) models, which needs to be implemented at some point.
	var failurePayload map[string]interface{}

	httpResponse, err := _sling.Receive(successPayload, &failurePayload)

	if err == nil {
		// err == nil only means that there wasn't a sub-application-layer error (e.g. no network error)
		if failurePayload != nil {
			// If the failurePayload is present, there likely was some kind of non-2xx status
			// returned (and a JSON payload error present)
			var str []byte
			str, err = json.Marshal(failurePayload)
			if err == nil { // For safety, check for an error marshalling... probably superfluous
				// This will return the JSON error body as a string
				err = errors.New(string(str))
			}
		} else {
			// So, there was no network-type error, and nothing in the failure payload,
			// but we should still check the status code
			if httpResponse == nil {
				// This should never happen...
				err = errors.New("No HTTP Response received.")
			} else if code := httpResponse.StatusCode; 200 > code || code > 299 {
				err = errors.New("HTTP Error: " + string(httpResponse.StatusCode))
			}
		}
	}

	return *successPayload, err
}

/**
 * Get a list of SSH keys.
 *
 * @param name SSH key name filter (exact match)
 * @param limit Number of results to display
 * @param offset The starting position of the query
 * @return SSHKeyList
 */
//func (a SshKeysApi) GetSSHKeyList (name string, limit int32, offset int32) (SSHKeyList, error) {
func (a SshKeysApi) GetSSHKeyList(name string, limit int32, offset int32) (SSHKeyList, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/sshKeys"

	_sling = _sling.Path(path)

	type QueryParams struct {
		name   string `url:"name,omitempty"`
		limit  int32  `url:"limit,omitempty"`
		offset int32  `url:"offset,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{name: name, limit: limit, offset: offset})
	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(SSHKeyList)

	// We use this map (below) so that any arbitrary error JSON can be handled.
	// FIXME: This is in the absence of this Go generator honoring the non-2xx
	// response (error) models, which needs to be implemented at some point.
	var failurePayload map[string]interface{}

	httpResponse, err := _sling.Receive(successPayload, &failurePayload)

	if err == nil {
		// err == nil only means that there wasn't a sub-application-layer error (e.g. no network error)
		if failurePayload != nil {
			// If the failurePayload is present, there likely was some kind of non-2xx status
			// returned (and a JSON payload error present)
			var str []byte
			str, err = json.Marshal(failurePayload)
			if err == nil { // For safety, check for an error marshalling... probably superfluous
				// This will return the JSON error body as a string
				err = errors.New(string(str))
			}
		} else {
			// So, there was no network-type error, and nothing in the failure payload,
			// but we should still check the status code
			if httpResponse == nil {
				// This should never happen...
				err = errors.New("No HTTP Response received.")
			} else if code := httpResponse.StatusCode; 200 > code || code > 299 {
				err = errors.New("HTTP Error: " + string(httpResponse.StatusCode))
			}
		}
	}

	return *successPayload, err
}
