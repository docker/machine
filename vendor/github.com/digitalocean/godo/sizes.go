package godo

// SizesService is an interface for interfacing with the size
// endpoints of the Digital Ocean API
// See: https://developers.digitalocean.com/#sizes
type SizesService interface {
	List(*ListOptions) ([]Size, *Response, error)
}

// SizesServiceOp handles communication with the size related methods of the
// DigitalOcean API.
type SizesServiceOp struct {
	client *Client
}

// Size represents a DigitalOcean Size
type Size struct {
	Slug         string   `json:"slug,omitempty"`
	Memory       int      `json:"memory,omitempty"`
	Vcpus        int      `json:"vcpus,omitempty"`
	Disk         int      `json:"disk,omitempty"`
	PriceMonthly float64  `json:"price_monthly,omitempty"`
	PriceHourly  float64  `json:"price_hourly,omitempty"`
	Regions      []string `json:"regions,omitempty"`
}

func (s Size) String() string {
	return Stringify(s)
}

type sizesRoot struct {
	Sizes []Size
	Links *Links `json:"links"`
}

// List all images
func (s *SizesServiceOp) List(opt *ListOptions) ([]Size, *Response, error) {
	path := "v2/sizes"
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(sizesRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Sizes, resp, err
}
