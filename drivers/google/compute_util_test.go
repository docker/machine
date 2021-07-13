package google

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	raw "google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
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

func TestOpenFirewallPorts(t *testing.T) {
	tests := map[string]struct {
		skipFirewall bool
		mockResponse http.HandlerFunc
	}{
		"skip firewall": {
			skipFirewall: true,
			mockResponse: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			}),
		},
		"firewall rules exists": {
			skipFirewall: false,
			mockResponse: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				firewall := raw.Firewall{

					Allowed: []*raw.FirewallAllowed{
						{
							IPProtocol: "tcp",
							Ports:      []string{"22", "2376"},
						},
					},
				}
				var body io.Reader = nil
				body, err := googleapi.WithoutDataWrapper.JSONReader(firewall)
				if err != nil {
					t.Fatal(err)
				}
				fmt.Fprint(w, body)
			}),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(tt.mockResponse)
			defer srv.Close()

			svc, err := raw.NewService(context.Background(), option.WithoutAuthentication(), option.WithEndpoint(srv.URL))
			if err != nil {
				t.Fatal(err)
			}

			computeUtil := ComputeUtil{
				skipFirewall: tt.skipFirewall,
				service:      svc,
			}

			driver := &Driver{}

			err = computeUtil.openFirewallPorts(driver)
			assert.NoError(t, err)
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

func TestPrepareMetadata(t *testing.T) {
	const (
		metadataKey1   = "key_1"
		metadataValue1 = "value_1"
		metadataKey2   = "key_2"
		metadataValue2 = "value_2"
		metadataKey3   = "key_3"
		fileContent3   = "file-content-3"
		metadataKey4   = "key_4"
		fileContent4   = "file-content-4"
	)

	metadata := metadataMap{
		metadataKey1: metadataValue1,
		metadataKey2: metadataValue2,
	}
	metadataFiles := [][]string{
		{metadataKey3, fileContent3},
		{metadataKey4, fileContent4},
	}

	noMetadataFile := func(_ *testing.T) (metadataMap, func()) { return metadataMap{}, func() {} }
	failingMetadataFile := func(_ *testing.T) (metadataMap, func()) {
		return metadataMap{"non-existing": "non-existing"}, func() {}
	}
	missingMetadataFilePath := func(_ *testing.T) (metadataMap, func()) {
		return metadataMap{"non-existing": ""}, func() {}
	}
	emptyMetadata := func(t *testing.T, m *raw.Metadata) {
		if !assert.NotNil(t, m) {
			t.FailNow()
		}
		assert.Empty(t, m.Items)
	}

	tests := map[string]struct {
		metadata       metadataMap
		metadataFiles  func(t *testing.T) (metadataMap, func())
		expectedError  bool
		assertMetadata func(t *testing.T, m *raw.Metadata)
	}{
		"error on metadata file reading": {
			metadataFiles:  failingMetadataFile,
			expectedError:  true,
			assertMetadata: emptyMetadata,
		},
		"missing metadata file path": {
			metadataFiles:  missingMetadataFilePath,
			expectedError:  false,
			assertMetadata: emptyMetadata,
		},
		"missing metadata value": {
			metadata:       metadataMap{"key": ""},
			metadataFiles:  noMetadataFile,
			expectedError:  false,
			assertMetadata: emptyMetadata,
		},
		"no metadata configuration": {
			metadata:       nil,
			metadataFiles:  noMetadataFile,
			expectedError:  false,
			assertMetadata: emptyMetadata,
		},
		"only metadata passed": {
			metadata:      metadata,
			metadataFiles: noMetadataFile,
			expectedError: false,
			assertMetadata: func(t *testing.T, m *raw.Metadata) {
				assertMetadata(t, m, metadataKey1, metadataValue1)
				assertMetadata(t, m, metadataKey2, metadataValue2)
			},
		},
		"only metadata file passed": {
			metadataFiles: prepareMetadataFiles(metadataFiles),
			expectedError: false,
			assertMetadata: func(t *testing.T, m *raw.Metadata) {
				assertMetadata(t, m, metadataKey3, fileContent3)
				assertMetadata(t, m, metadataKey4, fileContent4)
			},
		},
		"both metadata and metadata file passed": {
			metadata:      metadata,
			metadataFiles: prepareMetadataFiles(metadataFiles),
			expectedError: false,
			assertMetadata: func(t *testing.T, m *raw.Metadata) {
				assertMetadata(t, m, metadataKey1, metadataValue1)
				assertMetadata(t, m, metadataKey2, metadataValue2)
				assertMetadata(t, m, metadataKey3, fileContent3)
				assertMetadata(t, m, metadataKey4, fileContent4)
			},
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			metadataFiles, cleanup := tt.metadataFiles(t)
			defer cleanup()

			metadata, err := prepareMetadata(&Driver{
				Metadata:         tt.metadata,
				MetadataFromFile: metadataFiles,
			})

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			tt.assertMetadata(t, metadata)
		})
	}
}

func assertMetadata(t *testing.T, m *raw.Metadata, key string, value string) {
	if !assert.NotNil(t, m) {
		t.FailNow()
	}
	found := false
	for _, item := range m.Items {
		if item.Key != key {
			continue
		}

		found = true

		if !assert.NotNil(t, item.Value) {
			t.FailNow()
		}
		assert.Equal(t, value, *item.Value)
	}
	assert.True(t, found, "not found the metadata item %q=%q", key, value)
}

func prepareMetadataFiles(configuration [][]string) func(t *testing.T) (metadataMap, func()) {
	return func(t *testing.T) (metadataMap, func()) {
		metadata := make(metadataMap, 0)
		files := make([]string, 0)

		for _, entry := range configuration {
			file := prepareMetadataFile(t, entry[0], entry[1])

			files = append(files, file.Name())
			metadata[entry[0]] = file.Name()
		}

		cleanup := func() {
			for _, file := range files {
				err := os.RemoveAll(file)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			}
		}

		return metadata, cleanup
	}
}

func prepareMetadataFile(t *testing.T, key string, content string) *os.File {
	file, err := ioutil.TempFile("", "metadata-file-"+key)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	defer file.Close()

	_, err = fmt.Fprint(file, content)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return file
}

func TestAccelerator(t *testing.T) {
	tests := map[string]struct {
		description   string
		computeUtil   *ComputeUtil
		expectedCount int
		expectedType  string
	}{
		"unspecified": {
			computeUtil:   &ComputeUtil{},
			expectedCount: 0,
			expectedType:  "",
		},
		"GPU type": {
			computeUtil:   &ComputeUtil{accelerator: "type=nvidia-tesla-p100"},
			expectedCount: 1,
			expectedType:  "nvidia-tesla-p100",
		},
		"count and GPU type": {
			computeUtil:   &ComputeUtil{accelerator: "count=2,type=nvidia-tesla-p100"},
			expectedCount: 2,
			expectedType:  "nvidia-tesla-p100",
		},
		"count and GPU type with whitespace": {
			computeUtil:   &ComputeUtil{accelerator: " count=2, type=nvidia-tesla-p100 "},
			expectedCount: 2,
			expectedType:  "nvidia-tesla-p100",
		},
		"unknown key=value pair": {
			computeUtil:   &ComputeUtil{accelerator: "hello=world"},
			expectedCount: 0,
			expectedType:  "",
		},
		"extraneous key=value pair": {
			computeUtil:   &ComputeUtil{accelerator: "count=2,type=nvidia-tesla-p100,5"},
			expectedCount: 2,
			expectedType:  "nvidia-tesla-p100",
		},
		"invalid count": {
			computeUtil:   &ComputeUtil{accelerator: "count=ten,type=nvidia-tesla-p100"},
			expectedCount: 0,
			expectedType:  "",
		},
		"blank GPU type": {
			computeUtil:   &ComputeUtil{accelerator: "count=10,"},
			expectedCount: 10,
			expectedType:  "",
		},
	}

	for tn, tt := range tests {
		t.Run(tn, func(t *testing.T) {
			count, acceleratorType := tt.computeUtil.acceleratorCountAndType()

			assert.Equal(t, tt.expectedCount, count)
			assert.Equal(t, tt.expectedType, acceleratorType)
		})
	}
}
