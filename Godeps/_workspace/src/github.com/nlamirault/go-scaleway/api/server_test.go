// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"encoding/json"
	//"fmt"
	"testing"
)

func TestJsonServer(t *testing.T) {
	b := []byte(`{
   "server":{
      "tags":[
         "docker-machine"
      ],
      "state_detail":"",
      "image":{
         "default_bootscript":{
            "kernel":{
               "dtb":"dtb/pimouss-computing.dtb.3.17",
               "path":"kernel/pimouss-uImage-3.17-119-std",
               "id":"efff7963-2c2f-4467-837a-d14391218e36",
               "title":"Pimouss 3.17-119-std-with-aufs"
            },
            "title":"NBD Boot - Linux 3.17 119-std",
            "public":true,
            "initrd":{
               "path":"initrd/pimouss-uInitrd",
               "id":"fe70e4dc-fb87-47e8-bf61-4c75c6f5a61e",
               "title":"pimouss-uInitrd"
            },
            "bootcmdargs":{
               "id":"d22c4dde-e5a4-47ad-abb9-d23b54d542ff",
               "value":"ip=dhcp boot=local root=/dev/nbd0 USE_XNBD=1 nbd.max_parts=8"
            },
            "organization":"11111111-1111-4111-8111-111111111111",
            "id":"d28611ff-08bd-4bdd-9f73-084a0e1ec9dc"
         },
         "creation_date":"2015-01-19T18:08:41.906454+00:00",
         "name":"Debian Wheezy (7.8)",
         "modification_date":"2015-01-19T18:31:30.354525+00:00",
         "organization":"a283af0b-d13e-42e1-a43f-855ffbf281ab",
         "extra_volumes":"[]",
         "arch":"arm",
         "id":"cd66fa55-684a-4dd4-b809-956440b7a57f",
         "root_volume":{
            "size":20000000000,
            "id":"7f98d217-e7ec-4ce0-87ab-a76e9615400c",
            "volume_type":"l_ssd",
            "name":"distrib-debian-wheezy-2015-01-19_19:01-snapshot"
         },
         "public":true
      },
      "creation_date":"2015-01-21T10:00:24.619223+00:00",
      "public_ip":null,
      "private_ip":null,
      "id":"5f279dfb-6a92-4106-a0a6-21464a1fd6cf",
      "modification_date":"2015-01-21T10:00:24.619223+00:00",
      "name":"dev",
      "dynamic_public_ip":true,
      "hostname":"dev",
      "state":"stopped",
      "bootscript":null,
      "volumes":{
         "0":{
            "size":20000000000,
            "name":"distrib-debian-wheezy-2015-01-19_19:01-snapshot",
            "modification_date":"2015-01-21T09:56:45.041307+00:00",
            "organization":"19446e97-4a3b-4ccc-88f3-b65e3f31fb35",
            "export_uri":null,
            "creation_date":"2015-01-21T10:00:24.619223+00:00",
            "id":"6b6b97dc-f9d9-4916-80a5-b80a62c1c5ac",
            "volume_type":"l_ssd",
            "server":{
               "id":"5f279dfb-6a92-4106-a0a6-21464a1fd6cf",
               "name":"dev"
            }
         }
      },
      "organization":"19446e97-4a3b-4ccc-88f3-b65e3f31fb35"
   }
}`)
	var response ServerResponse
	err := json.Unmarshal(b, &response)
	if err != nil {
		t.Fatal(err)
	}
	name := "dev"
	if response.Server.Name != name {
		t.Fatalf("Expected %s, got %s",
			name, response.Server.Name)
	}
	organization := "19446e97-4a3b-4ccc-88f3-b65e3f31fb35"
	if response.Server.Organization != organization {
		t.Fatalf("Expected %s, got %s",
			organization, response.Server.Organization)
	}

}
