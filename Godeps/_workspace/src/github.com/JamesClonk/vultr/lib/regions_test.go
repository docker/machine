package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Regions_GetRegions_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	regions, err := client.GetRegions()
	assert.Nil(t, regions)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Regions_GetRegions_NoRegions(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	regions, err := client.GetRegions()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, regions)
}

func Test_Regions_GetRegions_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"5":{"DCID":"5","name":"Los Angeles","country":"US","continent":"North America","state":"CA"},
"9":{"DCID":"9","name":"Frankfurt","country":"DE","continent":"Europe","state":""},
"19":{"DCID":"19","name":"Australia","country":"AU","continent":"Australia","state":""}}`)
	defer server.Close()

	regions, err := client.GetRegions()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, regions) {
		assert.Equal(t, 3, len(regions))
		// Regions could be in random order
		for _, region := range regions {
			switch region.ID {
			case 5:
				assert.Equal(t, "Los Angeles", region.Name)
				assert.Equal(t, "US", region.Country)
				assert.Equal(t, "CA", region.State)
			case 9:
				assert.Equal(t, "Frankfurt", region.Name)
				assert.Equal(t, "DE", region.Country)
				assert.Equal(t, "Europe", region.Continent)
			case 19:
				assert.Equal(t, "AU", region.Country)
				assert.Equal(t, "", region.State)
				assert.Equal(t, "Australia", region.Continent)
			default:
				t.Error("Unknown DCID")
			}
		}
	}
}
