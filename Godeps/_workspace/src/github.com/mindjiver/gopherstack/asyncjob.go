package gopherstack

import (
	"net/url"
)

type QueryAsyncJobResultResponse struct {
	Queryasyncjobresultresponse struct {
		Accountid     string  `json:"accountid"`
		Cmd           string  `json:"cmd"`
		Created       string  `json:"created"`
		Jobid         string  `json:"jobid"`
		Jobprocstatus float64 `json:"jobprocstatus"`
		Jobresultcode float64 `json:"jobresultcode"`
		Jobstatus     float64 `json:"jobstatus"`
		Userid        string  `json:"userid"`
	} `json:"queryasyncjobresultresponse"`
}

// Query Cloudstack for the state of a scheduled job
func (c CloudstackClient) QueryAsyncJobResult(jobid string) (QueryAsyncJobResultResponse, error) {
	var resp QueryAsyncJobResultResponse

	params := url.Values{}
	params.Set("jobid", jobid)
	response, err := NewRequest(c, "queryAsyncJobResult", params)

	if err != nil {
		return resp, err
	}

	resp = response.(QueryAsyncJobResultResponse)

	return resp, nil
}
