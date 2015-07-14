package clcgo

const (
	apiDomain        = "https://api.tier3.com"
	apiRoot          = apiDomain + "/v2"
	successfulStatus = "succeeded"
)

// The Entity interface is implemented by any resource type that can be asked
// for its details via GetEntity. This seems basic and is implemented in most
// cases, but a few - PublicIPAddress for instance - can only be sent and not
// subsequently read.
type Entity interface {
	URL(string) (string, error)
}

type DeletionStatusProvidingEntity interface {
	StatusFromDeleteResponse([]byte) (Status, error)
}

// A Status is returned by all SaveEntity calls and can be used to determine
// when long-running provisioning jobs have completed. Things like Server
// creation take time, and you can periodically call GetEntity on the returned
// Status to determine if the server is ready.
type Status struct {
	Status string
	URI    string
}

func (s Status) URL(a string) (string, error) {
	return apiDomain + s.URI, nil
}

// HasSucceeded will, unsurprisingly, tell you if the Status is successful.
func (s Status) HasSucceeded() bool {
	return s.Status == successfulStatus
}

// A Link is a meta object returned in many API responses to help find
// resources related to the one you've requested.
type Link struct {
	ID   string `json:"id"`
	Rel  string `json:"rel"`
	HRef string `json:"href"`
}
