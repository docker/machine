package godo

// ImagesService is an interface for interfacing with the images
// endpoints of the Digital Ocean API
// See: https://developers.digitalocean.com/#images
type ImagesService interface {
	List(*ListOptions) ([]Image, *Response, error)
}

// ImagesServiceOp handles communication with the image related methods of the
// DigitalOcean API.
type ImagesServiceOp struct {
	client *Client
}

// Image represents a DigitalOcean Image
type Image struct {
	ID           int      `json:"id,float64,omitempty"`
	Name         string   `json:"name,omitempty"`
	Distribution string   `json:"distribution,omitempty"`
	Slug         string   `json:"slug,omitempty"`
	Public       bool     `json:"public,omitempty"`
	Regions      []string `json:"regions,omitempty"`
}

type imageRoot struct {
	Image Image
}

type imagesRoot struct {
	Images []Image
	Links  *Links `json:"links"`
}

func (i Image) String() string {
	return Stringify(i)
}

// List all sizes
func (s *ImagesServiceOp) List(opt *ListOptions) ([]Image, *Response, error) {
	path := "v2/images"
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(imagesRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Images, resp, err
}
