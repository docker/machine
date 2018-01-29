package egoscale

import (
	"testing"
)

func TestSnapshots(t *testing.T) {
	var _ AsyncCommand = (*CreateSnapshot)(nil)
	var _ Command = (*ListSnapshots)(nil)
	var _ AsyncCommand = (*DeleteSnapshot)(nil)
	var _ AsyncCommand = (*RevertSnapshot)(nil)
}

func TestCreateSnapshot(t *testing.T) {
	req := &CreateSnapshot{}
	if req.name() != "createSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*CreateSnapshotResponse)
}

func TestListSnapshots(t *testing.T) {
	req := &ListSnapshots{}
	if req.name() != "listSnapshots" {
		t.Errorf("API call doesn't match")
	}
	_ = req.response().(*ListSnapshotsResponse)
}

func TestDeleteSnapshot(t *testing.T) {
	req := &DeleteSnapshot{}
	if req.name() != "deleteSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}

func TestRevertSnapshot(t *testing.T) {
	req := &RevertSnapshot{}
	if req.name() != "revertSnapshot" {
		t.Errorf("API call doesn't match")
	}
	_ = req.asyncResponse().(*booleanAsyncResponse)
}
