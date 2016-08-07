package cloud

type ServerList struct {
	Results    []Server   `json:"results,omitempty"`
	Pagination Pagination `json:"pagination,omitempty"`
}
