package azureutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/libmachine/mcnutils"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/rancher/machine/drivers/azure/logutil"
)

const (
	validateAuthorizerTimeout = time.Second * 5
)

// Azure driver allows two authentication methods:
//
// 1. OAuth Device Flow
//
// Azure Active Directory implements OAuth 2.0 Device Flow described here:
// https://tools.ietf.org/html/draft-denniss-oauth-device-flow-00. It is simple
// for users to authenticate through a browser and requires re-authenticating
// every 2 weeks.
//
// Device auth prints a message to the screen telling the user to click on URL
// and approve the app on the browser, meanwhile the client polls the auth API
// for a token. Once we have token, we save it locally to a file with proper
// permissions and when the token expires (in Azure case typically 1 hour) SDK
// will automatically refresh the specified token and will call the refresh
// callback function we implement here. This way we will always be storing a
// token with a refresh_token saved on the machine.
//
// 2. Azure Service Principal Account
//
// This is designed for headless authentication to Azure APIs but requires more
// steps from user to create a Service Principal Account and provide its
// credentials to the machine driver.

var (
	// AD app id for docker-machine driver in various Azure realms
	appIDs = map[string]string{
		azure.PublicCloud.Name: "637ddaba-219b-43b8-bf19-8cea500cf273",
		azure.ChinaCloud.Name:  "bb5eed6f-120b-4365-8fd9-ab1a3fba5698",
		azure.GermanCloud.Name: "aabac5f7-dd47-47ef-824c-e0d57598cada",
	}
)

// AuthenticateDeviceFlow fetches a token from the local file cache or initiates a consent
// flow and waits for token to be obtained. Obtained token is stored in a file cache for
// future use and refreshing.
func AuthenticateDeviceFlow(ctx context.Context, env azure.Environment, subscriptionID string) (*autorest.BearerAuthorizer, error) {
	clientID, ok := appIDs[env.Name]
	if !ok {
		return nil, fmt.Errorf("docker-machine application not set up for Azure environment %q", env.Name)
	}
	// We locate the tenant ID of the subscription as we store tokens per
	// tenant (which could have multiple subscriptions)
	tenantID, err := loadOrFindTenantID(ctx, env, subscriptionID)
	if err != nil {
		return nil, err
	}
	tokenPath := tokenCachePath(tenantID)
	if err != nil {
		return nil, err
	}
	deviceFlowConfig := auth.DeviceFlowConfig{
		ClientID:    clientID,
		TenantID:    tenantID,
		AADEndpoint: env.ActiveDirectoryEndpoint,
		Resource:    env.ResourceManagerEndpoint,
	}

	servicePrincipalToken, err := loadToken(tokenPath, deviceFlowConfig)
	if err != nil {
		// Unexpected failure at loadToken
		return nil, err
	}
	if servicePrincipalToken != nil {
		log.Debug("Auth token found in file.", logutil.Fields{"path": tokenPath})
		if err := servicePrincipalToken.EnsureFresh(); err != nil {
			return nil, err
		}
		// Invoke saving the token only if it was refreshed
		if err := servicePrincipalToken.InvokeRefreshCallbacks(servicePrincipalToken.Token()); err != nil {
			return nil, err
		}
	} else {
		// Otherwise, generate a new service principal token and save it
		servicePrincipalToken, err = deviceFlowConfig.ServicePrincipalToken()
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokenPath, servicePrincipalToken.Token()); err != nil {
			return nil, err
		}
	}
	authorizer := autorest.NewBearerAuthorizer(servicePrincipalToken)
	ValidateAuthorizer(ctx, env, authorizer)
	return authorizer, nil
}

// AuthenticateClientCredentials uses given client credentials to return a
// service principal token. Generated token is not stored in a cache file or refreshed.
func AuthenticateClientCredentials(ctx context.Context, env azure.Environment, subscriptionID, clientID, clientSecret string) (*autorest.BearerAuthorizer, error) {
	tenantID, err := loadOrFindTenantID(ctx, env, subscriptionID)
	if err != nil {
		return nil, err
	}
	clientCredentialsConfig := auth.ClientCredentialsConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
		AADEndpoint:  env.ActiveDirectoryEndpoint,
		Resource:     env.ResourceManagerEndpoint,
	}
	servicePrincipalToken, err := clientCredentialsConfig.ServicePrincipalToken()
	if err != nil {
		return nil, err
	}
	authorizer := autorest.NewBearerAuthorizer(servicePrincipalToken)
	ValidateAuthorizer(ctx, env, authorizer)
	return authorizer, nil
}

// ValidateAuthorizer makes a call to Azure SDK with given authorizer to make sure it is valid
func ValidateAuthorizer(ctx context.Context, env azure.Environment, authorizer *autorest.BearerAuthorizer) error {
	goCtx, cancel := context.WithTimeout(ctx, validateAuthorizerTimeout)
	defer cancel()
	c := subscriptionsClient(env.ResourceManagerEndpoint)
	c.Authorizer = authorizer
	_, err := c.List(goCtx)
	if err != nil {
		return fmt.Errorf("Authorizer validity check failed: %v", err)
	}
	return nil
}

// loadToken returns a token from the specified file if it is found, otherwise
// returns nil. Any error retrieving or creating the token is returned as an error.
func loadToken(tokenPath string, dfc auth.DeviceFlowConfig) (*adal.ServicePrincipalToken, error) {
	log.Debug("Loading auth token from file", logutil.Fields{"path": tokenPath})
	if _, err := os.Stat(tokenPath); err != nil {
		if os.IsNotExist(err) { // file not found
			return nil, nil
		}
		return nil, err
	}
	token, err := adal.LoadToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load token from file: %v", err)
	}
	oauthCfg, err := adal.NewOAuthConfig(dfc.AADEndpoint, dfc.TenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}
	saveTokenCallback := func(t adal.Token) error {
		log.Debug("Azure token expired. Saving the refreshed token...")
		return saveToken(tokenPath, t)
	}

	servicePrincipalToken, err := adal.NewServicePrincipalTokenFromManualToken(*oauthCfg, dfc.ClientID, dfc.Resource, *token, saveTokenCallback)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return servicePrincipalToken, nil
}

func saveToken(tokenPath string, token adal.Token) error {
	if err := adal.SaveToken(tokenPath, 0600, token); err != nil {
		log.Error("Error occurred saving token to cache file.")
		return err
	}
	log.Debug("Saved token to file", logutil.Fields{"path": tokenPath})
	return nil
}

// azureCredsPath returns the directory the azure credentials are stored in.
func azureCredsPath() string {
	return filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine", "credentials", "azure")
}

// tokenCachePath returns the full path the OAuth 2.0 token should be saved at
// for given tenant ID.
func tokenCachePath(tenantID string) string {
	return filepath.Join(azureCredsPath(), fmt.Sprintf("%s.json", tenantID))
}
