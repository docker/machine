package godo

// ActionRequest reprents DigitalOcean Action Request
type ActionRequest struct {
	Type   string                 `json:"type"`
	Params map[string]interface{} `json:"params,omitempty"`
}

// Converts an ActionRequest to a string.
func (d ActionRequest) String() string {
	return Stringify(d)
}
