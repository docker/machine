package egoscale

import (
	"testing"
)

func TestEvents(t *testing.T) {
	var _ Command = (*ListEvents)(nil)
	var _ Command = (*ListEventTypes)(nil)
}

func TestListEvents(t *testing.T) {
	req := &ListEvents{}
	if req.name() != "listEvents" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListEventsResponse)
}

func TestListEventTypes(t *testing.T) {
	req := &ListEventTypes{}
	if req.name() != "listEventTypes" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListEventTypesResponse)
}
