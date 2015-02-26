package linodego

import (
	log "github.com/Sirupsen/logrus"
	"testing"
)

func TestListImages(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Image.List()
	if err != nil {
		t.Fatal(err)
	}

	for _, image := range v.Images {
		log.Debugf("Kernel: %s, %s, %s", image.Label, image.Status, image.CreateDt)
	}
}

func TestUpdateImage(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Image.Update(94057, "", "Mongodb Image 2")
	if err != nil {
		t.Fatal(err)
	}

	image := v.Image
	log.Debugf("Kernel: %s, %s, %s", image.Label, image.Status, image.CreateDt)
}
