package linodego

import (
	"encoding/json"
	"net/url"
	"strconv"
)

// Job service
type LinodeJobService struct {
	client *Client
}

// Resonse for linode.job.list API
type LinodesJobListResponse struct {
	Response
	Jobs []Job
}

// List all jobs. If jobId is greater than 0, limit the list to given jobId.
func (t *LinodeJobService) List(linodeId int, jobId int, pendingOnly bool) (*LinodesJobListResponse, error) {
	u := &url.Values{}
	u.Add("LinodeID", strconv.Itoa(linodeId))
	if pendingOnly {
		u.Add("pendingOnly", "1")
	}
	if jobId > 0 {
		u.Add("JobID", strconv.Itoa(jobId))
	}
	v := LinodesJobListResponse{}
	if err := t.client.do("linode.job.list", u, &v.Response); err != nil {
		return nil, err
	}

	v.Jobs = make([]Job, 5)
	if err := json.Unmarshal(v.RawData, &v.Jobs); err != nil {
		return nil, err
	}
	return &v, nil
}
