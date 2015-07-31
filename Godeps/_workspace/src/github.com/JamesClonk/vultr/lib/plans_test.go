package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Plans_GetPlans_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	plans, err := client.GetPlans()
	assert.Nil(t, plans)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Plans_GetPlans_NoPlans(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	plans, err := client.GetPlans()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, plans)
}

func Test_Plans_GetPlans_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"29":{"VPSPLANID":"29","name":"768 MB RAM,15 GB SSD,1.00 TB BW","vcpu_count":"1","ram":"768","disk":"15","bandwidth":"1.00","bandwidth_gb":"1024","price_per_month":"5.00","windows":false},
"30":{"VPSPLANID":"30","name":"1024 MB RAM,20 GB SSD,2.00 TB BW","vcpu_count":"2","ram":"1024","disk":"20","bandwidth":"2.00","bandwidth_gb":"2048","price_per_month":"7.00","windows":false}}`)
	defer server.Close()

	plans, err := client.GetPlans()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, plans) {
		assert.Equal(t, 2, len(plans))
		// Plans could be in random order
		for _, plan := range plans {
			switch plan.ID {
			case 29:
				assert.Equal(t, "768 MB RAM,15 GB SSD,1.00 TB BW", plan.Name)
				assert.Equal(t, 1, plan.VCpus)
				assert.Equal(t, "768", plan.RAM)
				assert.Equal(t, "5.00", plan.Price)
			case 30:
				assert.Equal(t, "1024 MB RAM,20 GB SSD,2.00 TB BW", plan.Name)
				assert.Equal(t, 2, plan.VCpus)
				assert.Equal(t, "20", plan.Disk)
				assert.Equal(t, "2.00", plan.Bandwidth)
			default:
				t.Error("Unknown VPSPLANID")
			}
		}
	}
}

func Test_Plans_GetAvailablePlansForRegion_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	plans, err := client.GetAvailablePlansForRegion(1)
	assert.Nil(t, plans)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_Plans_GetAvailablePlansForRegion_NoPlans(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	plans, err := client.GetAvailablePlansForRegion(2)
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, plans)
}

func Test_Plans_GetAvailablePlansForRegion_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[29,30,3,27,28,11,13,81]`)
	defer server.Close()

	plans, err := client.GetAvailablePlansForRegion(3)
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, plans) {
		assert.Equal(t, 8, len(plans))
	}
}
