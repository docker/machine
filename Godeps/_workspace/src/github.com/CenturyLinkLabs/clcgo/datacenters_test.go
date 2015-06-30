package clcgo

import (
	"encoding/json"
	"testing"

	"github.com/CenturyLinkLabs/clcgo/fakeapi"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulDataCentersURL(t *testing.T) {
	d := DataCenters{}
	u, err := d.URL("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/datacenters/AA", u)
}

func TestDataCenterJSONUnmarshalling(t *testing.T) {
	j := `{"id": "foo", "name": "bar"}`

	dc := DataCenter{}
	err := json.Unmarshal([]byte(j), &dc)

	assert.NoError(t, err)

	assert.Equal(t, "foo", dc.ID)

	assert.Equal(t, "bar", dc.Name)
}

func TestSuccessfulDataCenterCapabilitiesURL(t *testing.T) {
	d := DataCenterCapabilities{DataCenter: DataCenter{ID: "abc123"}}
	u, err := d.URL("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/datacenters/AA/abc123/deploymentCapabilities", u)
}

func TestErroredDataCenterCapabilitiesURL(t *testing.T) {
	d := DataCenterCapabilities{}
	s, err := d.URL("AA")
	assert.Equal(t, "", s)
	assert.EqualError(t, err, "need a DataCenter with an ID")
}

func TestSuccessfulDataCenterCapabilitiesUnmarshalling(t *testing.T) {
	d := DataCenterCapabilities{}
	err := json.Unmarshal([]byte(fakeapi.DataCenterCapabilitiesResponse), &d)

	assert.NoError(t, err)
	if assert.Len(t, d.Templates, 1) {
		assert.Equal(t, "Name", d.Templates[0].Name)
		assert.Equal(t, "Description", d.Templates[0].Description)
	}

	if assert.Len(t, d.DeployableNetworks, 1) {
		assert.Equal(t, "Test Network", d.DeployableNetworks[0].Name)
		assert.Equal(t, "id-for-network", d.DeployableNetworks[0].NetworkID)
	}
}

func TestSuccessfulDataCenterGroupURL(t *testing.T) {
	d := DataCenterGroup{DataCenter: DataCenter{ID: "abc123"}}
	u, err := d.URL("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/datacenters/AA/abc123?groupLinks=true", u)
}

func TestErroredDataCenterGroupURL(t *testing.T) {
	d := DataCenterGroup{}
	_, err := d.URL("AA")
	assert.EqualError(t, err, "need a DataCenter with an ID")
}

func TestSuccessfulDataCenterGroupUnmarshalling(t *testing.T) {
	d := DataCenterGroup{}
	err := json.Unmarshal([]byte(fakeapi.DataCenterGroupResponse), &d)

	assert.NoError(t, err)
	assert.Equal(t, "IL1", d.ID)
	assert.Equal(t, "Illinois 1", d.Name)
	assert.Len(t, d.Links, 2)
}
