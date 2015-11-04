package godo

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestImages_ImagesServiceOpImplementsImagesService(t *testing.T) {
	if !Implements((*ImagesService)(nil), new(ImagesServiceOp)) {
		t.Error("ImagesServiceOp does not implement ImagesService")
	}
}

func TestImages_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"images":[{"id":1},{"id":2}]}`)
	})

	images, _, err := client.Images.List(nil)
	if err != nil {
		t.Errorf("Images.List returned error: %v", err)
	}

	expected := []Image{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(images, expected) {
		t.Errorf("Images.List returned %+v, expected %+v", images, expected)
	}
}

func TestImages_ListImagesMultiplePages(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"images": [{"id":1},{"id":2}], "links":{"pages":{"next":"http://example.com/v2/images/?page=2"}}}`)
	})

	_, resp, err := client.Images.List(&ListOptions{Page: 2})
	if err != nil {
		t.Fatal(err)
	}
	checkCurrentPage(t, resp, 1)
}

func TestImages_RetrievePageByNumber(t *testing.T) {
	setup()
	defer teardown()

	jBlob := `
	{
		"images": [{"id":1},{"id":2}],
		"links":{
			"pages":{
				"next":"http://example.com/v2/images/?page=3",
				"prev":"http://example.com/v2/images/?page=1",
				"last":"http://example.com/v2/images/?page=3",
				"first":"http://example.com/v2/images/?page=1"
			}
		}
	}`

	mux.HandleFunc("/v2/images", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, jBlob)
	})

	opt := &ListOptions{Page: 2}
	_, resp, err := client.Images.List(opt)
	if err != nil {
		t.Fatal(err)
	}

	checkCurrentPage(t, resp, 2)
}

func TestImage_String(t *testing.T) {
	image := &Image{
		ID:           1,
		Name:         "Image",
		Distribution: "Ubuntu",
		Slug:         "image",
		Public:       true,
		Regions:      []string{"one", "two"},
	}

	stringified := image.String()
	expected := `godo.Image{ID:1, Name:"Image", Distribution:"Ubuntu", Slug:"image", Public:true, Regions:["one" "two"]}`
	if expected != stringified {
		t.Errorf("Image.String returned %+v, expected %+v", stringified, expected)
	}
}
