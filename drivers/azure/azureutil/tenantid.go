package azureutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/docker/machine/libmachine/log"

	"github.com/Azure/go-autorest/autorest/azure"
)

// laodOrFindTenantID figures out the AAD tenant ID of the subscription by first
// looking at the cache file, if not exists, makes a network call to load it and
// cache it for future use.
func loadOrFindTenantID(env azure.Environment, subscriptionID string) (string, error) {
	var tenantID string

	// Load from cache
	fp := tenantIDPath(subscriptionID)
	b, err := ioutil.ReadFile(fp)
	if err == nil {
		tenantID = strings.TrimSpace(string(b))
		log.Debugf("Tenant ID loaded from file: %s", fp)
	} else {
		if os.IsNotExist(err) {
			log.Debugf("Tenant ID file not found: %s", fp)
		} else {
			return "", fmt.Errorf("Failed to load tenant ID file: %v", err)
		}
	}

	// Handle cache miss
	if tenantID == "" {
		log.Debug("Making API call to find tenant ID")
		t, err := findTenantID(env, subscriptionID)
		if err != nil {
			return "", err
		}
		tenantID = t

		// Cache the result
		if err := saveTenantID(fp, tenantID); err != nil {
			return "", fmt.Errorf("Failed to save tenant ID: %v", err)
		}
		log.Debugf("Cached tenant ID to file: %s", fp)
	}
	return tenantID, nil
}

// findTenantID figures out the AAD tenant ID of the subscription by making an
// unauthenticated request to the Get Subscription Details endpoint and parses
// the value from WWW-Authenticate header.
func findTenantID(env azure.Environment, subscriptionID string) (string, error) {
	const hdrKey = "WWW-Authenticate"
	c := subscriptionsClient(env.ResourceManagerEndpoint)

	// we expect this request to fail (err != nil), but we are only interested
	// in headers, so surface the error if the Response is not present (i.e.
	// network error etc)
	subs, err := c.Get(subscriptionID)
	if subs.Response.Response == nil {
		return "", fmt.Errorf("Request failed: %v", err)
	}

	// Expecting 401 StatusUnauthorized here, just read the header
	if subs.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("Unexpected response from Get Subscription: %v", err)
	}
	hdr := subs.Header.Get(hdrKey)
	if hdr == "" {
		return "", fmt.Errorf("Header %v not found in Get Subscription response", hdrKey)
	}

	// Example value for hdr:
	//   Bearer authorization_uri="https://login.windows.net/996fe9d1-6171-40aa-945b-4c64b63bf655", error="invalid_token", error_description="The authentication failed because of missing 'Authorization' header."
	r := regexp.MustCompile(`authorization_uri=".*/([0-9a-f\-]+)"`)
	m := r.FindStringSubmatch(hdr)
	if m == nil {
		return "", fmt.Errorf("Could not find the tenant ID in header: %s %q", hdrKey, hdr)
	}
	return m[1], nil
}

func saveTenantID(path string, tenantID string) error {
	return ioutil.WriteFile(path, []byte(tenantID), 0600)
}
