package cloud

type Pagination struct {
	Limit int32  `json:"limit,omitempty"`
	Prev  string `json:"prev,omitempty"`
	Next  string `json:"next,omitempty"`
	Total int64  `json:"total,omitempty"`
}
