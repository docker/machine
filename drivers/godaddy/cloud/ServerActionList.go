package cloud

type ServerActionList struct {
	Results    []ServerAction `json:"results,omitempty"`
	Pagination Pagination     `json:"pagination,omitempty"`
}
