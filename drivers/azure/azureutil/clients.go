package azureutil

import (
	"fmt"
	"time"

	"github.com/rancher/machine/version"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-11-01/subscriptions"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
)

const (
	defaultClientPollingDelay = time.Second * 5
)

// TODO(ahmetalpbalkan) Remove duplication around client creation. This is
// happening because we auto-generate our SDK and we don't have generics in Go.
// We are hoping to come up with a factory or some defaults instance to set
// these client configuration in a central place in azure-sdk-for-go.

func oauthClient() autorest.Client {
	c := autorest.NewClientWithUserAgent(fmt.Sprintf("docker-machine/%s", version.Version))
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	// TODO set user agent
	return c
}

func subscriptionsClient(baseURI string) subscriptions.Client {
	c := subscriptions.NewClientWithBaseURI(baseURI) // used only for unauthenticated requests for generic subs IDs
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) providersClient() resources.ProvidersClient {
	c := resources.NewProvidersClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) resourceGroupsClient() resources.GroupsClient {
	c := resources.NewGroupsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) securityGroupsClient() network.SecurityGroupsClient {
	c := network.NewSecurityGroupsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) virtualNetworksClient() network.VirtualNetworksClient {
	c := network.NewVirtualNetworksClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) subnetsClient() network.SubnetsClient {
	c := network.NewSubnetsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) networkInterfacesClient() network.InterfacesClient {
	c := network.NewInterfacesClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) publicIPAddressClient() network.PublicIPAddressesClient {
	c := network.NewPublicIPAddressesClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) storageAccountsClient() storage.AccountsClient {
	c := storage.NewAccountsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) blobContainersClient() storage.BlobContainersClient {
	c := storage.NewBlobContainersClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) virtualMachinesClient() compute.VirtualMachinesClient {
	c := compute.NewVirtualMachinesClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) availabilitySetsClient() compute.AvailabilitySetsClient {
	c := compute.NewAvailabilitySetsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) imagesClient() compute.ImagesClient {
	c := compute.NewImagesClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) galleryImageVersionsClient() compute.GalleryImageVersionsClient {
	c := compute.NewGalleryImageVersionsClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) resourceSkusClient() compute.ResourceSkusClient {
	c := compute.NewResourceSkusClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}

func (a AzureClient) disksClient() compute.DisksClient {
	c := compute.NewDisksClientWithBaseURI(a.env.ResourceManagerEndpoint, a.subscriptionID)
	c.Authorizer = a.auth
	c.Client.UserAgent += fmt.Sprintf(";docker-machine/%s", version.Version)
	c.RequestInspector = withInspection()
	c.ResponseInspector = byInspecting()
	c.PollingDelay = defaultClientPollingDelay
	return c
}
