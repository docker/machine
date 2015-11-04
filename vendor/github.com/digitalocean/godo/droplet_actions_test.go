package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestDropletActions_DropletActionsServiceOpImplementsDropletActionsService(t *testing.T) {
	if !Implements((*DropletActionsService)(nil), new(DropletActionsServiceOp)) {
		t.Error("DropletActionsServiceOp does not implement DropletActionsService")
	}
}

func TestDropletActions_Shutdown(t *testing.T) {
	setup()
	defer teardown()

	request := &ActionRequest{
		Type: "shutdown",
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.DropletActions.Shutdown(1)
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_PowerOff(t *testing.T) {
	setup()
	defer teardown()

	request := &ActionRequest{
		Type: "power_off",
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.DropletActions.PowerOff(1)
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_PowerOn(t *testing.T) {
	setup()
	defer teardown()

	request := &ActionRequest{
		Type: "power_on",
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.DropletActions.PowerOn(1)
	if err != nil {
		t.Errorf("DropletActions.PowerOn returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.PowerOn returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_Reboot(t *testing.T) {
	setup()
	defer teardown()

	request := &ActionRequest{
		Type: "reboot",
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	action, _, err := client.DropletActions.Reboot(1)
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_Restore(t *testing.T) {
	setup()
	defer teardown()

	options := map[string]interface{}{
		"image": float64(1),
	}

	request := &ActionRequest{
		Type:   "restore",
		Params: options,
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")

		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	action, _, err := client.DropletActions.Restore(1, 1)
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_Resize(t *testing.T) {
	setup()
	defer teardown()

	options := map[string]interface{}{
		"size": "1024mb",
	}

	request := &ActionRequest{
		Type:   "resize",
		Params: options,
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")

		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	action, _, err := client.DropletActions.Resize(1, "1024mb")
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_Rename(t *testing.T) {
	setup()
	defer teardown()

	options := map[string]interface{}{
		"name": "Droplet-Name",
	}

	request := &ActionRequest{
		Type:   "rename",
		Params: options,
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")

		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.DropletActions.Rename(1, "Droplet-Name")
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletAction_PowerCycle(t *testing.T) {
	setup()
	defer teardown()

	request := &ActionRequest{
		Type: "power_cycle",
	}

	mux.HandleFunc("/v2/droplets/1/actions", func(w http.ResponseWriter, r *http.Request) {
		v := new(ActionRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		if !reflect.DeepEqual(v, request) {
			t.Errorf("Request body = %+v, expected %+v", v, request)
		}

		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)

	})

	action, _, err := client.DropletActions.PowerCycle(1)
	if err != nil {
		t.Errorf("DropletActions.Shutdown returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Shutdown returned %+v, expected %+v", action, expected)
	}
}

func TestDropletActions_Get(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/123/actions/456", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{"action":{"status":"in-progress"}}`)
	})

	action, _, err := client.DropletActions.Get(123, 456)
	if err != nil {
		t.Errorf("DropletActions.Get returned error: %v", err)
	}

	expected := &Action{Status: "in-progress"}
	if !reflect.DeepEqual(action, expected) {
		t.Errorf("DropletActions.Get returned %+v, expected %+v", action, expected)
	}
}
