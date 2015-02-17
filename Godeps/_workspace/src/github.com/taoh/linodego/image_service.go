package linodego

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Image Service
type ImageService struct {
	client *Client
}

// Response for image.list API
type ImagesListResponse struct {
	Response
	Images []Image
}

// Response for image.update or image.delete API
type ImageResponse struct {
	Response
	Image Image
}

// List all images
func (t *ImageService) List() (*ImagesListResponse, error) {
	u := &url.Values{}
	v := ImagesListResponse{}
	if err := t.client.do("image.list", u, &v.Response); err != nil {
		return nil, err
	}

	v.Images = make([]Image, 5)
	if err := json.Unmarshal(v.RawData, &v.Images); err != nil {
		return nil, err
	}
	return &v, nil
}

// Update given Image
func (t *ImageService) Update(imageId int, label string, description string) (*ImageResponse, error) {
	u := &url.Values{}
	u.Add("ImageID", strconv.Itoa(imageId))
	if label != "" {
		u.Add("label", label)
	}
	if description != "" {
		u.Add("description", description)
	}

	v := ImageResponse{}
	if err := t.client.do("image.update", u, &v.Response); err != nil {
		return nil, err
	}

	v.Image = Image{}
	if err := json.Unmarshal(v.RawData, &v.Image); err != nil {
		return nil, err
	}
	return &v, nil
}

// Delete given Image
func (t *ImageService) Delete(imageId int) (*ImageResponse, error) {
	u := &url.Values{}
	u.Add("ImageID", strconv.Itoa(imageId))
	v := ImageResponse{}
	if err := t.client.do("image.delete", u, &v.Response); err != nil {
		return nil, err
	}

	v.Image = Image{}
	if err := json.Unmarshal(v.RawData, &v.Image); err != nil {
		return nil, err
	}
	return &v, nil
}
