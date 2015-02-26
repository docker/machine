package linodego

import (
	log "github.com/Sirupsen/logrus"
	"testing"
)

func TestAvailDataCenters(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Avail.DataCenters()
	if err != nil {
		t.Fatal(err)
	}

	for _, dataCenter := range v.DataCenters {
		log.Debugf("Data Center: %d, %s, %s", dataCenter.DataCenterId, dataCenter.Location, dataCenter.Abbr)
	}
}

func TestAvailDistributions(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Avail.Distributions()
	if err != nil {
		t.Fatal(err)
	}

	for _, distribution := range v.Distributions {
		log.Debugf("Distribution: %s, %b, %s", distribution.Label, distribution.Is64Bit, distribution.CreatedDt)
	}
}

func TestAvailKernels(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Avail.Kernels()
	if err != nil {
		t.Fatal(err)
	}

	for _, kernel := range v.Kernels {
		log.Debugf("Kernel: %s, %d, %d, %d", kernel.Label, kernel.IsXen, kernel.IsPVOPS, kernel.KernelId)
	}
}

func TestAvailLinodePlans(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Avail.LinodePlans()
	if err != nil {
		t.Fatal(err)
	}

	for _, plan := range v.LinodePlans {
		log.Debugf("Linode Plans: %d, %s, %f, %d, %d", plan.PlanId, plan.Label, plan.Price, plan.Cores, plan.RAM)
	}
}

func TestAvailNodeBalancers(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Avail.NodeBalancers()
	if err != nil {
		t.Fatal(err)
	}

	for _, nodeBalancer := range v.NodeBalancers {
		log.Debugf("NodeBalancer: %f, %f, %d", nodeBalancer.Hourly, nodeBalancer.Monthly, nodeBalancer.Connections)
	}
}

// func TestAvailStackScripts(t *testing.T) {
// 	client := NewClient(APIKey, nil)

// 	v, err := client.Avail.StackScripts()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// for _, stackScript := range v.StackScripts {
// 	// 	log.Debugf("StackScripts: %s, %d, %d", stackScript.Label, stackScript.DeploymentsTotal, stackScript.DeploymentsActive)
// 	// }
// }
