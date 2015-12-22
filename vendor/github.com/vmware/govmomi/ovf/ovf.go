/*
Copyright (c) 2015 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ovf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

func Unmarshal(r io.Reader) (*Envelope, error) {
	var e Envelope

	dec := xml.NewDecoder(r)
	err := dec.Decode(&e)
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func Marshal(e Env) (string, error) {
	x, err := xml.Marshal(e)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", xml.Header, x), nil
}

const (
	ovfEnvHeader = `<Environment
		xmlns="http://schemas.dmtf.org/ovf/environment/1"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		xmlns:oe="http://schemas.dmtf.org/ovf/environment/1"
		xmlns:ve="http://www.vmware.com/schema/ovfenv"
		oe:id=""
		ve:esxId="%s">`
	ovfEnvPlatformSection = `<PlatformSection>
		<Kind>%s</Kind>
		<Version>%s</Version>
		<Vendor>%s</Vendor>
		<Locale>%s</Locale>
		</PlatformSection>`
	ovfEnvPropertyHeader = `<PropertySection>`
	ovfEnvPropertyEntry  = `<Property oe:key="%s" oe:value="%s"/>`
	ovfEnvPropertyFooter = `</PropertySection>`
	ovfEnvFooter         = `</Environment>`
)

// MarshalManual manually marshals Env to xml suitable for a vApp guest.
// It exists to overcome the lack of expressiveness in golang's xml namespace.
func MarshalManual(e Env) string {
	var buffer bytes.Buffer

	buffer.WriteString(xml.Header)
	buffer.WriteString(fmt.Sprintf(ovfEnvHeader, e.EsxID))
	buffer.WriteString(fmt.Sprintf(ovfEnvPlatformSection, e.Platform.Kind, e.Platform.Version, e.Platform.Vendor, e.Platform.Locale))

	buffer.WriteString(fmt.Sprintf(ovfEnvPropertyHeader))
	for _, p := range e.Property.Properties {
		buffer.WriteString(fmt.Sprintf(ovfEnvPropertyEntry, p.Key, p.Value))
	}
	buffer.WriteString(fmt.Sprintf(ovfEnvPropertyFooter))

	buffer.WriteString(fmt.Sprintf(ovfEnvFooter))

	return buffer.String()
}
