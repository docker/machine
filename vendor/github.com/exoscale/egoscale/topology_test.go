package egoscale

import (
	"log"
	"testing"
)

func TestGetImages(t *testing.T) {
	ts := newServer(200, `
{
	"listtemplatesresponse (doesn't matter)": {
		"count": 0,
		"template": [
			{
				"id": "4c0732a0-3df0-4f66-8d16-009f91cf05d6",
				"name": "Linux RedHat 7.4 64-bit",
				"displayText": "Linux RedHat 7.4 64-bit 10G Disk (2017-11-31-dummy)",
				"size": 10737418240
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux Ubuntu 12.04 LTS 64-bit",
				"displayText": "Linux Ubuntu 12.04 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux Debian 8 64-bit",
				"displayText": "Linux Debian 8 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux CentOS 7.3 64-bit",
				"displayText": "Linux CentOS 7.3 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			},{
				"id": "1959ccb7-cd79-404d-a156-322e4a0c3beb",
				"name": "Linux CoreOS stable 1298 64-bit",
				"displayText": "Linux CoreOS stable 1298 64-bit 50G Disk (2017-11-31-dummy)",
				"size": 53687091200
			}
		]
	}
}
	`)
	defer ts.Close()

	cs := NewClient(ts.URL, "TOKEN", "SECRET")
	images, err := cs.GetImages()
	if err != nil {
		log.Fatal(err)
	}

	var tests = []struct {
		uuid  string
		names []string
		size  int64
	}{
		{
			"4c0732a0-3df0-4f66-8d16-009f91cf05d6",
			[]string{"redhat-7.4", "linux redhat 7.4 64-bit"},
			10,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"ubuntu-12.04", "linux ubuntu 12.04 lts 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"debian-8", "linux debian 8 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"centos-7.3", "linux centos 7.3 64-bit"},
			50,
		}, {
			"1959ccb7-cd79-404d-a156-322e4a0c3beb",
			[]string{"coreos-stable-1298", "linux coreos stable 1298 64-bit"},
			50,
		},
	}

	for _, test := range tests {
		for _, name := range test.names {
			if _, ok := images[name]; !ok {
				t.Errorf("expected %s into the map", name)
			}

			if _, ok := images[name][test.size]; !ok {
				t.Errorf("expected %s, %dG into the map", name, test.size)
			}

			if uuid := images[name][test.size]; uuid != test.uuid {
				t.Errorf("bad uuid for the %s image. got %v expected %v", name, uuid, test.uuid)
			}
		}
	}
}
