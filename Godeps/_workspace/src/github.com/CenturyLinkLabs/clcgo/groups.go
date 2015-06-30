package clcgo

import (
	"errors"
	"fmt"
)

const (
	groupsURL = apiRoot + "/groups/%s"
	groupURL  = groupsURL + "/%s"
)

// A Group resource can be used to discover the hierarchy of the created groups
// for your account for a given datacenter. The Group's ID can either be for
// one you have created, or for a datacenter's hardware group (which can be
// determined via the DataCenterGroup resource).
//
// When creating a new group, the ParentGroup field must be set before you
// attempt to save.
type Group struct {
	ParentGroup   *Group  `json:"-"`
	ID            string  `json:"id"`
	ParentGroupID string  `json:"parentGroupId"` // TODO: not in get, extract to creation params?
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	Type          string  `json:"type"`
	Groups        []Group `json:"groups"`
}

func (g *Group) URL(a string) (string, error) {
	if g.ID == "" {
		return "", errors.New("an ID field is required to get a group")
	}

	return fmt.Sprintf(groupURL, a, g.ID), nil
}

func (g *Group) RequestForSave(a string) (request, error) {
	if g.ParentGroup == nil || g.ParentGroup.ID == "" {
		return request{}, errors.New("a ParentGroup with an ID is required to create a group")
	}

	url := fmt.Sprintf(groupsURL, a)
	g.ParentGroupID = g.ParentGroup.ID
	return request{URL: url, Parameters: *g}, nil
}
