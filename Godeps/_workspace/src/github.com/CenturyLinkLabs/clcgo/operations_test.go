package clcgo

import (
	"testing"

	"github.com/CenturyLinkLabs/clcgo/fakeapi"
	"github.com/stretchr/testify/assert"
)

func TestSuccessfulServerOperationRequestForSave(t *testing.T) {
	s := Server{ID: "test-id"}
	p := ServerOperation{Server: s, OperationType: PauseServer}
	req, err := p.RequestForSave("AA")

	assert.NoError(t, err)
	assert.Equal(t, apiRoot+"/operations/AA/servers/pause", req.URL)

	sids, ok := req.Parameters.([]string)
	if assert.True(t, ok) {
		assert.Len(t, sids, 1)
		assert.Equal(t, "test-id", sids[0])
	}
}

func TestErroredServerOperationRequestForSave(t *testing.T) {
	p := ServerOperation{OperationType: PauseServer}
	req, err := p.RequestForSave("AA")

	assert.Equal(t, request{}, req)
	assert.EqualError(t, err, "a ServerOperation requires a Server and OperationType")

	s := Server{ID: "test-id"}
	p = ServerOperation{Server: s}
	req, err = p.RequestForSave("AA")

	assert.Equal(t, request{}, req)
	assert.EqualError(t, err, "a ServerOperation requires a Server and OperationType")
}

func TestServerOperationStatusFromCreateResponse(t *testing.T) {
	s := Server{ID: "test-id"}
	p := ServerOperation{Server: s, OperationType: PauseServer}
	st, err := p.StatusFromCreateResponse([]byte(fakeapi.PauseServersSuccessfulResponse))
	assert.NoError(t, err)
	assert.Equal(t, "/path/to/status", st.URI)
}
