package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestImageActions_ImageActionsServiceOpImplementsImageActionsService(t *testing.T) {
	if !Implements((*ImageActionsService)(nil), new(ImageActionsServiceOp)) {
		t.Error("ImageActionsServiceOp does not implement ImageActionsService")
	}
}

func TestImageActions_Transfer(t *testing.T) {
	setup()
	defer teardown()

	transferRequest := &ActionRequest{}

	mux.HandleFunc("/v2/images/12345/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, transferRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, transferRequest)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	transfer, _, err := client.ImageActions.Transfer(12345, transferRequest)
	if err != nil {
		t.Errorf("ImageActions.Transfer returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(transfer, expected) {
		t.Errorf("ImageActions.Transfer returned %+v, expected %+v", transfer, expected)
	}
}

func TestImageActions_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/images/123/actions/456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.ImageActions.Get(123, 456)
	if err != nil {
		t.Errorf("ImageActions.Get returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("ImageActions.Get returned %+v, expected %+v", action, expected)
	}
}
