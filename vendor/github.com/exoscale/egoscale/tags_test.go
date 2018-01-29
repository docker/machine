package egoscale

import (
	"testing"
)

func TestTags(t *testing.T) {
	var _ AsyncCommand = (*CreateTags)(nil)
	var _ AsyncCommand = (*DeleteTags)(nil)
	var _ Command = (*ListTags)(nil)
}

func TestCreateTags(t *testing.T) {
	req := &CreateTags{}
	if req.name() != "createTags" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestDeleteTags(t *testing.T) {
	req := &DeleteTags{}
	if req.name() != "deleteTags" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestListTags(t *testing.T) {
	req := &ListTags{}
	if req.name() != "listTags" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListTagsResponse)
}
