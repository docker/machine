package linodego

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Account Service
type AccountService struct {
	client *Client
}

// Response for account.estimateinvoice API
type EstimateInvoiceResponse struct {
	Response
	EstimateInvoice EstimateInvoice
}

// Response for account.info API
type AccountInfoResponse struct {
	Response
	AccountInfo AccountInfo
}

// Estimate Invoice
func (t *AccountService) EstimateInvoice(mode string, planId int, paymentTerm int, linodeId int) (*EstimateInvoiceResponse, error) {
	u := &url.Values{}
	u.Add("mode", mode)
	u.Add("PlanId", strconv.Itoa(planId))
	u.Add("LinodeId", strconv.Itoa(linodeId))

	if (mode == "linode_new") || (mode == "nodebalancer_new") {
		u.Add("PaymentTerm", strconv.Itoa(paymentTerm))
	}
	//TODO: add more validations for params combinations
	v := EstimateInvoiceResponse{}
	if err := t.client.do("account.estimateinvoice", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.EstimateInvoice); err != nil {
		return nil, err
	}
	return &v, nil
}

// Get Account Info
func (t *AccountService) Info() (*AccountInfoResponse, error) {
	u := &url.Values{}
	v := AccountInfoResponse{}
	if err := t.client.do("account.info", u, &v.Response); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(v.RawData, &v.AccountInfo); err != nil {
		return nil, err
	}
	return &v, nil
}
