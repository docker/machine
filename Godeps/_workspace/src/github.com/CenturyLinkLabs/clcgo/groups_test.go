package clcgo

import (
	"encoding/json"
	"testing"

	"github.com/CenturyLinkLabs/clcgo/fakeapi"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulGroupURL(t *testing.T) {
	g := Group{ID: "test-id"}
	u, err := g.URL("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/groups/AA/test-id", u)
}

func TestErroredGroupURL(t *testing.T) {
	g := Group{}
	s, err := g.URL("AA")
	assert.Equal(t, "", s)
	assert.EqualError(t, err, "an ID field is required to get a group")
}

func TestGroupUnmarshalling(t *testing.T) {
	g := Group{}
	err := json.Unmarshal([]byte(fakeapi.GroupResponse), &g)

	assert.NoError(t, err)
	assert.Equal(t, "test-id", g.ID)
	assert.Equal(t, "Test Group", g.Name)
	assert.Equal(t, "archive", g.Type)
	if assert.Len(t, g.Groups, 1) {
		assert.Equal(t, "test-child-id", g.Groups[0].ID)
	}
}

func TestSuccessfulGroupRequestForSave(t *testing.T) {
	p := Group{ID: "parent-group-id"}
	g := Group{
		Name:        "Test Group",
		Description: "Description",
		ParentGroup: &p,
	}
	req, err := g.RequestForSave("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/groups/AA", req.URL)
	assert.Equal(t, g, req.Parameters)
	rg, ok := req.Parameters.(Group)
	if assert.True(t, ok) {
		assert.Equal(t, "parent-group-id", rg.ParentGroupID)
	}
}

func TestErroredGroupRequestForSave(t *testing.T) {
	g := Group{Name: "Test Group"}
	req, err := g.RequestForSave("AA")
	assert.Equal(t, request{}, req)
	assert.EqualError(t, err, "a ParentGroup with an ID is required to create a group")
}
