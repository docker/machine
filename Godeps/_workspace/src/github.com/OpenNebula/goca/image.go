package goca

import (
	"errors"
	"strconv"
)

type Image struct {
	XMLResource
	Id   uint
	Name string
}

type ImagePool struct {
	XMLResource
}

type IMAGE_STATE int

const (
	IMAGE_INIT IMAGE_STATE = iota
	IMAGE_READY
	IMAGE_USED
	IMAGE_DISABLED
	IMAGE_LOCKED
	IMAGE_ERROR
	IMAGE_CLONE
	IMAGE_DELETE
	IMAGE_USED_PERS
)

func (s IMAGE_STATE) String() string {
	return [...]string{
		"INIT",
		"READY",
		"USED",
		"DISABLED",
		"LOCKED",
		"ERROR",
		"CLONE",
		"DELETE",
		"USED_PERS",
	}[s]
}

func CreateImage(template string, ds_id uint) (uint, error) {
	response, err := client.Call("one.image.allocate", template, ds_id)
	if err != nil {
		return 0, err
	}

	return uint(response.BodyInt()), nil
}

func NewImagePool(args ...int) (*ImagePool, error) {
	var who, start_id, end_id, state int

	switch len(args) {
	case 0:
		who = PoolWhoMine
		start_id = -1
		end_id = -1
		state = -1
	case 3:
		who = args[0]
		start_id = args[1]
		end_id = args[2]
	default:
		return nil, errors.New("Wrong number of arguments")
	}

	response, err := client.Call("one.imagepool.info", who, start_id, end_id, state)
	if err != nil {
		return nil, err
	}

	vmpool := &ImagePool{XMLResource{body: response.Body()}}

	return vmpool, err

}

func NewImage(id uint) *Image {
	return &Image{Id: id}
}

func NewImageFromName(name string) (*Image, error) {
	imagePool, err := NewImagePool()
	if err != nil {
		return nil, err
	}

	id, err := imagePool.GetIdFromName(name, "/IMAGE_POOL/IMAGE")
	if err != nil {
		return nil, err
	}

	return NewImage(id), nil
}

func (image *Image) Info() error {
	response, err := client.Call("one.image.info", image.Id)
	image.body = response.Body()
	return err
}

func (image *Image) State() (int, error) {
	stateString, ok := image.XPath("/IMAGE/STATE")
	if ok != true {
		return -1, errors.New("Unable to parse Image State")
	}

	state, _ := strconv.Atoi(stateString)

	return state, nil
}

func (image *Image) StateString() (string, error) {
	image_state, err := image.State()
	if err != nil {
		return "", err
	}
	return IMAGE_STATE(image_state).String(), nil
}

func (image *Image) Delete() error {
	_, err := client.Call("one.image.delete", image.Id)
	return err
}
