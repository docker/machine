// +build acceptance identity

package v2

import (
	"testing"

	tokens2 "github.com/rackspace/gophercloud/openstack/identity/v2/tokens"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestAuthenticate(t *testing.T) {
	ao := v2AuthOptions(t)
	service := unauthenticatedClient(t)

	// Authenticated!
	result := tokens2.Create(service, tokens2.WrapOptions(ao))

	// Extract and print the token.
	token, err := result.ExtractToken()
	th.AssertNoErr(t, err)

	t.Logf("Acquired token: [%s]", token.ID)
	t.Logf("The token will expire at: [%s]", token.ExpiresAt.String())
	t.Logf("The token is valid for tenant: [%#v]", token.Tenant)

	// Extract and print the service catalog.
	catalog, err := result.ExtractServiceCatalog()
	th.AssertNoErr(t, err)

	t.Logf("Acquired service catalog listing [%d] services", len(catalog.Entries))
	for i, entry := range catalog.Entries {
		t.Logf("[%02d]: name=[%s], type=[%s]", i, entry.Name, entry.Type)
		for _, endpoint := range entry.Endpoints {
			t.Logf("      - region=[%s] publicURL=[%s]", endpoint.Region, endpoint.PublicURL)
		}
	}
}
