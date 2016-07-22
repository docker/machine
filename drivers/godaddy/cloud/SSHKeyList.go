package cloud

type SSHKeyList struct {
	Results    []SSHKey   `json:"results,omitempty"`
	Pagination Pagination `json:"pagination,omitempty"`
}
