package cloud

type ModelError struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}
