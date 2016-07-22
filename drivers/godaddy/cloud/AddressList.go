package cloud

type AddressList struct {
	Results    []IpAddress `json:"results,omitempty"`
	Pagination Pagination  `json:"pagination,omitempty"`
}
