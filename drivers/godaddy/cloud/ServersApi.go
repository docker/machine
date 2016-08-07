package cloud

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dghubble/sling"
)

type ServersApi struct {
	basePath string
	sling    *sling.Sling
}

func NewServersApi(basePath string, apiKey string) *ServersApi {
	s := sling.New().Set("Authorization", fmt.Sprintf("sso-key %s", apiKey))
	return &ServersApi{
		basePath: basePath,
		sling:    s,
	}
}

/**
 * Create a new server
 * Use to initiate the provisioning process for a new server
 * @param body server details
 * @return Server
 */
//func (a ServersApi) AddServer (body ServerCreate) (Server, error) {
func (a ServersApi) AddServer(body ServerCreate) (Server, error) {

	_sling := a.sling.New().Post(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers"

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	// body params
	_sling = _sling.BodyJSON(body)

	var successPayload = new(Server)

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
 * Get a console URL to this server
 * This URL can be viewed in a web browser, and allows you to access your server even when its network is down because of misconfigured interfaces or iptables rules, or when it is failing to boot properly.
 * @param serverId Server to access (serverId)
 * @return Console
 */
//func (a ServersApi) Console (serverId string) (Console, error) {
func (a ServersApi) Console(serverId string) (Console, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/console"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(Console)

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
 * Destroy an existing server
 * Use to initiate the shutdown and destruction of an existing server.
 * @param serverId Id of server to be destroyed
 * @return ServerAction
 */
//func (a ServersApi) DestroyServer (serverId string) (ServerAction, error) {
func (a ServersApi) DestroyServer(serverId string) (ServerAction, error) {

	_sling := a.sling.New().Post(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/destroy"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerAction)

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
 * Get ServerAction by id
 *
 * @param serverId serverId of associated server
 * @param serverActionId Id of ServerAction to be fetched
 * @return ServerAction
 */
//func (a ServersApi) GetServerActionById (serverId string, serverActionId string) (ServerAction, error) {
func (a ServersApi) GetServerActionById(serverId string, serverActionId string) (ServerAction, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/actions/{serverActionId}"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)
	path = strings.Replace(path, "{"+"serverActionId"+"}", fmt.Sprintf("%v", serverActionId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerAction)

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
 * List of actions performed on the specified server
 *
 * @param serverId serverId of associated server
 * @param type_ Action type filter (exact match)
 * @param limit Number of results to display
 * @param offset The starting position of the query
 * @return ServerActionList
 */
//func (a ServersApi) GetServerActionList (serverId string, type_ string, limit int32, offset int32) (ServerActionList, error) {
func (a ServersApi) GetServerActionList(serverId string, type_ string, limit int32, offset int32) (ServerActionList, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/actions"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	type QueryParams struct {
		type_  string `url:"type,omitempty"`
		limit  int32  `url:"limit,omitempty"`
		offset int32  `url:"offset,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{type_: type_, limit: limit, offset: offset})
	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerActionList)

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
 * Find Addresses by serverId and addressId
 *
 * @param serverId serverId of associated server
 * @param addressId Id of Address to be fetched
 * @return IpAddress
 */
//func (a ServersApi) GetServerAddressById (serverId string, addressId string) (IpAddress, error) {
func (a ServersApi) GetServerAddressById(serverId string, addressId string) (IpAddress, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/addresses/{addressId}"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)
	path = strings.Replace(path, "{"+"addressId"+"}", fmt.Sprintf("%v", addressId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(IpAddress)

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
 * List of Addresses of the specified server
 *
 * @param serverId serverId of associated server
 * @param address Numeric address (exact match)
 * @param status Address status (exact match)
 * @param type_ Address type (exact match)
 * @param limit Number of results to display
 * @param offset The starting position of the query
 * @return AddressList
 */
//func (a ServersApi) GetServerAddressList (serverId string, address string, status string, type_ string, limit int32, offset int32) (AddressList, error) {
func (a ServersApi) GetServerAddressList(serverId string, address string, status string, type_ string, limit int32, offset int32) (AddressList, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/addresses"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	type QueryParams struct {
		address string `url:"address,omitempty"`
		status  string `url:"status,omitempty"`
		type_   string `url:"type,omitempty"`
		limit   int32  `url:"limit,omitempty"`
		offset  int32  `url:"offset,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{address: address, status: status, type_: type_, limit: limit, offset: offset})
	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(AddressList)

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
 * Find server by serverId
 *
 * @param serverId Id of server to be fetched
 * @return Server
 */
//func (a ServersApi) GetServerById (serverId string) (Server, error) {
func (a ServersApi) GetServerById(serverId string) (Server, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(Server)

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
 * Get a list of servers. By default, all destroyed servers are filtered out.
 *
 * @param status Server status filter (exact match)
 * @param backupsEnabled BackupsEnabled flag
 * @param limit Number of results to display
 * @param offset The starting position of the query
 * @return ServerList
 */
//func (a ServersApi) GetServerList (status string, backupsEnabled bool, limit int32, offset int32) (ServerList, error) {
func (a ServersApi) GetServerList(status string, backupsEnabled bool, limit int32, offset int32) (ServerList, error) {

	_sling := a.sling.New().Get(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers"

	_sling = _sling.Path(path)

	type QueryParams struct {
		status         string `url:"status,omitempty"`
		backupsEnabled bool   `url:"backupsEnabled,omitempty"`
		limit          int32  `url:"limit,omitempty"`
		offset         int32  `url:"offset,omitempty"`
	}
	_sling = _sling.QueryStruct(&QueryParams{status: status, backupsEnabled: backupsEnabled, limit: limit, offset: offset})
	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerList)

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
 * Update server details
 *
 * @param serverId serverId of server to be updated
 * @param body Server data
 * @return Server
 */
//func (a ServersApi) PatchServer (serverId string, body Server) (Server, error) {
func (a ServersApi) PatchServer(serverId string, body Server) (Server, error) {

	_sling := a.sling.New().Patch(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	// body params
	_sling = _sling.BodyJSON(body)

	var successPayload = new(Server)

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
 * Start a server
 * Use to start a stopped server
 * @param serverId serverId of server to be started
 * @return ServerAction
 */
//func (a ServersApi) StartServer (serverId string) (ServerAction, error) {
func (a ServersApi) StartServer(serverId string) (ServerAction, error) {

	_sling := a.sling.New().Post(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/start"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerAction)

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
 * Stop a server
 * Use to stop a running server
 * @param serverId serverId of server to be stopped
 * @return ServerAction
 */
//func (a ServersApi) StopServer (serverId string) (ServerAction, error) {
func (a ServersApi) StopServer(serverId string) (ServerAction, error) {

	_sling := a.sling.New().Post(a.basePath)

	// create path and map variables
	path := "/v1/cloud/servers/{serverId}/stop"
	path = strings.Replace(path, "{"+"serverId"+"}", fmt.Sprintf("%v", serverId), -1)

	_sling = _sling.Path(path)

	// accept header
	accepts := []string{"application/json"}
	for key := range accepts {
		_sling = _sling.Set("Accept", accepts[key])
		break // only use the first Accept
	}

	var successPayload = new(ServerAction)

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
