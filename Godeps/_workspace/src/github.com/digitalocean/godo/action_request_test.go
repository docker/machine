package godo

import "testing"

func TestActionRequest_String(t *testing.T) {
	action := &ActionRequest{
		Type:   "transfer",
		Params: map[string]interface{}{"key-1": "value-1"},
	}

	stringified := action.String()
	expected := `godo.ActionRequest{Type:"transfer", Params:map[key-1:value-1]}`
	if expected != stringified {
		t.Errorf("Action.Stringify returned %+v, expected %+v", stringified, expected)
	}
}
