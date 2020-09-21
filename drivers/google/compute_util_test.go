package google

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	raw "google.golang.org/api/compute/v1"
)

func TestDefaultTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: ""})

	assert.Equal(t, []string{"docker-machine"}, tags)
}

func TestAdditionalTag(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1"})

	assert.Equal(t, []string{"docker-machine", "tag1"}, tags)
}

func TestAdditionalTags(t *testing.T) {
	tags := parseTags(&Driver{Tags: "tag1,tag2"})

	assert.Equal(t, []string{"docker-machine", "tag1", "tag2"}, tags)
}

func TestLabels(t *testing.T) {
	tests := map[string]struct {
		labels         []string
		expectedLabels map[string]string
	}{
		"no labels": {
			labels:         []string{},
			expectedLabels: map[string]string{},
		},
		"1 label": {
			labels: []string{"key1:value1"},
			expectedLabels: map[string]string{
				"key1": "value1",
			},
		},
		"multiple label": {
			labels: []string{"key1: value1", "key2: value2"},
			expectedLabels: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		"label missing key": {
			labels: []string{"value1"},
			expectedLabels: map[string]string{
				"value1": "",
			},
		},
		"label missing value": {
			labels: []string{"key1:"},
			expectedLabels: map[string]string{
				"key1": "",
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			labels := parseLabels(&Driver{Labels: tt.labels})

			assert.Equal(t, tt.expectedLabels, labels)
		})
	}
}

func TestPortsUsed(t *testing.T) {
	var tests = []struct {
		description   string
		computeUtil   *ComputeUtil
		expectedPorts []string
		expectedError error
	}{
		{"use docker port", &ComputeUtil{}, []string{"2376/tcp"}, nil},
		{"use docker and swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:3376"}, []string{"2376/tcp", "3376/tcp"}, nil},
		{"use docker and non default swarm port", &ComputeUtil{SwarmMaster: true, SwarmHost: "tcp://host:4242"}, []string{"2376/tcp", "4242/tcp"}, nil},
		{"include additional ports", &ComputeUtil{openPorts: []string{"80", "2377/udp"}}, []string{"2376/tcp", "80/tcp", "2377/udp"}, nil},
	}

	for _, test := range tests {
		ports, err := test.computeUtil.portsUsed()

		assert.Equal(t, test.expectedPorts, ports)
		assert.Equal(t, test.expectedError, err)
	}
}

func TestMissingOpenedPorts(t *testing.T) {
	var tests = []struct {
		description     string
		allowed         []*raw.FirewallAllowed
		ports           []string
		expectedMissing map[string][]string
	}{
		{"no port opened", []*raw.FirewallAllowed{}, []string{"2376"}, map[string][]string{"tcp": {"2376"}}},
		{"docker port opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376"}, map[string][]string{}},
		{"missing swarm port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"3376"}}},
		{"missing docker port", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"3376"}}}, []string{"2376", "3376"}, map[string][]string{"tcp": {"2376"}}},
		{"both ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"more ports opened", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "3376", "22", "1024-2048"}}}, []string{"2376", "3376"}, map[string][]string{}},
		{"additional missing", []*raw.FirewallAllowed{{IPProtocol: "tcp", Ports: []string{"2376", "2377/tcp"}}}, []string{"2377/udp", "80/tcp", "2376"}, map[string][]string{"tcp": {"80"}, "udp": {"2377"}}},
	}

	for _, test := range tests {
		firewall := &raw.Firewall{Allowed: test.allowed}

		missingPorts := missingOpenedPorts(firewall, test.ports)

		assert.Equal(t, test.expectedMissing, missingPorts, test.description)
	}
}

type testOperationCaller struct {
	operationDuration time.Duration
	getError          error
	operationError    *raw.OperationErrorErrors

	calls     int
	startedAt time.Time
}

func (oc *testOperationCaller) Get() (*raw.Operation, error) {
	oc.calls++

	if oc.getError != nil {
		return nil, oc.getError
	}

	op := &raw.Operation{
		Name: "test operation",
	}

	if time.Since(oc.startedAt) >= oc.operationDuration {
		op.Status = "DONE"
	} else {
		op.Status = "PENDING"
	}

	if oc.operationError != nil {
		op.Error = &raw.OperationError{
			Errors: []*raw.OperationErrorErrors{
				oc.operationError,
			},
		}
	}

	return op, nil
}

func TestWaitForOpBackOff(t *testing.T) {
	tests := map[string]struct {
		backoffFactoryNotDefined bool
		operationDuration        time.Duration
		maxOperationDuration     time.Duration

		getError       error
		operationError *raw.OperationErrorErrors

		expectedError error
	}{
		"error on call": {
			getError:      errors.New("test error"),
			expectedError: errors.New("test error"),
		},
		"operation too long": {
			operationDuration:    5 * time.Second,
			maxOperationDuration: 1 * time.Second,
			expectedError:        errors.New("maximum backoff elapsed time exceeded"),
		},
		"operation error": {
			operationDuration:    1 * time.Second,
			maxOperationDuration: 5 * time.Second,
			operationError: &raw.OperationErrorErrors{
				Code:     "code",
				Location: "location",
				Message:  "message",
			},
			expectedError: errors.New("operation error: {code location message [] []}"),
		},
		"backoff factory not defined": {
			backoffFactoryNotDefined: true,
			operationDuration:        5 * time.Second,
			maxOperationDuration:     5 * time.Second,
			expectedError:            errors.New("operationBackoffFactory is not defined"),
		},
		"proper operation call": {
			operationDuration:    5 * time.Second,
			maxOperationDuration: 5 * time.Second,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			toc := &testOperationCaller{
				operationDuration: test.operationDuration,
				getError:          test.getError,
				operationError:    test.operationError,

				startedAt: time.Now(),
			}

			cu := &ComputeUtil{}
			if !test.backoffFactoryNotDefined {
				cu.operationBackoffFactory = &backoffFactory{
					InitialInterval:     125 * time.Millisecond,
					RandomizationFactor: 0,
					Multiplier:          2,
					MaxInterval:         4 * time.Second,
					MaxElapsedTime:      test.maxOperationDuration,
				}
			}

			err := cu.waitForOp(toc.Get)
			if test.expectedError != nil {
				assert.EqualError(t, err, test.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, toc.calls < 8, "Too many *OperationServices.Get() calls")
		})
	}
}
