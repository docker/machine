package linodego

import (
	log "github.com/Sirupsen/logrus"
	"testing"
)

func TestAccountEstimateInvoice(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Account.EstimateInvoice("linode_new", 2, 1, -1)
	if err != nil {
		t.Fatal(err)
	}

	log.Debugf("Estimate Invoice: %s, %f", v.EstimateInvoice.InvoiceTo, v.EstimateInvoice.Amount)
}

func TestAccountInfo(t *testing.T) {
	client := NewClient(APIKey, nil)

	v, err := client.Account.Info()
	if err != nil {
		t.Fatal(err)
	}

	log.Debugf("Account Info: %f, %b", v.AccountInfo.Balance, v.AccountInfo.Managed)
}
